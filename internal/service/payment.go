package service

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type PaymentService struct {
	repo          ports.PaymentRepository
	subscriptions ports.SubscriptionRepository
	plans         ports.PlanRepository
	periods       ports.BillingPeriodRepository
	balances      ports.SubscriptionBalanceRepository
	allocations   ports.PaymentAllocationRepository
	audit         ports.AuditRepository
	txRunner      ports.PaymentTxRunner
	now           func() time.Time
}

func NewPaymentService(
	repo ports.PaymentRepository,
	subscriptions ports.SubscriptionRepository,
	plans ports.PlanRepository,
	periods ports.BillingPeriodRepository,
	balances ports.SubscriptionBalanceRepository,
	allocations ports.PaymentAllocationRepository,
	audit ports.AuditRepository,
	txRunner ports.PaymentTxRunner,
) *PaymentService {
	return &PaymentService{
		repo:          repo,
		subscriptions: subscriptions,
		plans:         plans,
		periods:       periods,
		balances:      balances,
		allocations:   allocations,
		audit:         audit,
		txRunner:      txRunner,
		now:           time.Now,
	}
}

func (s *PaymentService) Register(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	if s.txRunner != nil {
		var result domain.Payment
		err := s.txRunner.RunSerializable(ctx, func(ctx context.Context, deps ports.PaymentDependencies) error {
			txService := &PaymentService{
				repo:          deps.Payments,
				subscriptions: deps.Subscriptions,
				plans:         deps.Plans,
				periods:       deps.BillingPeriods,
				balances:      deps.Balances,
				allocations:   deps.Allocations,
				audit:         deps.Audit,
				now:           s.now,
			}
			var err error
			result, err = txService.register(ctx, payment)
			return err
		})
		if err != nil {
			return domain.Payment{}, err
		}
		return result, nil
	}

	return s.register(ctx, payment)
}

func (s *PaymentService) register(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	metadata := map[string]any{
		"amount_cents":    payment.AmountCents,
		"method":          string(payment.Method),
		"subscription_id": payment.SubscriptionID,
		"idempotency_key": payment.IdempotencyKey,
	}
	recordAuditAttempt(ctx, s.audit, "payment.create", "payment", payment.ID, metadata)

	if payment.SubscriptionID == "" {
		err := errors.New("assinatura e obrigatoria")
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}
	if payment.AmountCents <= 0 {
		err := errors.New("valor deve ser maior que zero")
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}

	if s.subscriptions == nil || s.plans == nil || s.periods == nil || s.allocations == nil {
		err := errors.New("dependencias de pagamentos indisponiveis")
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}

	if payment.IdempotencyKey != "" {
		existing, err := s.repo.FindByIdempotencyKey(ctx, payment.IdempotencyKey)
		if err == nil {
			idemMetadata := copyMetadata(metadata)
			idemMetadata["idempotent"] = true
			recordAuditSuccess(ctx, s.audit, "payment.create", "payment", existing.ID, idemMetadata)
			return existing, nil
		}
		if err != nil && !errors.Is(err, ports.ErrNotFound) {
			recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
			return domain.Payment{}, err
		}
	}

	subscription, err := s.subscriptions.FindByID(ctx, payment.SubscriptionID)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}
	if subscription.Status != domain.SubscriptionActive {
		err := errors.New("assinatura inativa")
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}

	plan, err := s.plans.FindByID(ctx, subscription.PlanID)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}

	if payment.Status == "" {
		payment.Status = domain.PaymentConfirmed
	}
	if payment.PaidAt.IsZero() {
		payment.PaidAt = s.now()
	}
	if payment.Kind == "" {
		payment.Kind = domain.PaymentFull
	}
	if !payment.Method.IsValid() {
		err := errors.New("metodo de pagamento invalido")
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}
	if !payment.Status.IsValid() {
		err := errors.New("status de pagamento invalido")
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}
	if !payment.Kind.IsValid() {
		err := errors.New("tipo de pagamento invalido")
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}

	today := dateOnly(s.now())
	if _, err := s.ensurePeriods(ctx, subscription, plan, today); err != nil {
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}

	created, err := s.repo.Create(ctx, payment)
	if err != nil {
		if errors.Is(err, ports.ErrConflict) && payment.IdempotencyKey != "" {
			existing, findErr := s.repo.FindByIdempotencyKey(ctx, payment.IdempotencyKey)
			if findErr == nil {
				idemMetadata := copyMetadata(metadata)
				idemMetadata["idempotent"] = true
				recordAuditSuccess(ctx, s.audit, "payment.create", "payment", existing.ID, idemMetadata)
				return existing, nil
			}
			return domain.Payment{}, findErr
		}
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", payment.ID, metadata, err)
		return domain.Payment{}, err
	}

	result, err := s.applyPayment(ctx, created, subscription, today)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", created.ID, metadata, err)
		return created, err
	}

	created.Kind = result.Kind
	created.CreditCents = result.CreditCents
	updated, err := s.repo.Update(ctx, created)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "payment.create", "payment", created.ID, metadata, err)
		return created, err
	}

	successMetadata := copyMetadata(metadata)
	successMetadata["credit_cents"] = updated.CreditCents
	successMetadata["kind"] = string(updated.Kind)
	successMetadata["status"] = string(updated.Status)
	recordAuditSuccess(ctx, s.audit, "payment.create", "payment", updated.ID, successMetadata)
	return updated, nil
}

func (s *PaymentService) Update(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	if payment.ID == "" {
		err := errors.New("pagamento invalido")
		recordAuditFailure(ctx, s.audit, "payment.update", "payment", payment.ID, nil, err)
		return domain.Payment{}, err
	}
	recordAuditAttempt(ctx, s.audit, "payment.update", "payment", payment.ID, nil)

	current, err := s.repo.FindByID(ctx, payment.ID)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "payment.update", "payment", payment.ID, nil, err)
		return domain.Payment{}, err
	}

	if payment.SubscriptionID == "" {
		payment.SubscriptionID = current.SubscriptionID
	}
	if payment.Method == "" {
		payment.Method = current.Method
	}
	if payment.AmountCents == 0 {
		payment.AmountCents = current.AmountCents
	}
	if payment.PaidAt.IsZero() {
		payment.PaidAt = current.PaidAt
	}
	if payment.Status == "" {
		payment.Status = current.Status
	}
	payment.Kind = current.Kind
	payment.CreditCents = current.CreditCents
	payment.IdempotencyKey = current.IdempotencyKey
	if !payment.Method.IsValid() {
		err := errors.New("metodo de pagamento invalido")
		recordAuditFailure(ctx, s.audit, "payment.update", "payment", payment.ID, nil, err)
		return domain.Payment{}, err
	}
	if !payment.Status.IsValid() {
		err := errors.New("status de pagamento invalido")
		recordAuditFailure(ctx, s.audit, "payment.update", "payment", payment.ID, nil, err)
		return domain.Payment{}, err
	}
	if !payment.Kind.IsValid() {
		err := errors.New("tipo de pagamento invalido")
		recordAuditFailure(ctx, s.audit, "payment.update", "payment", payment.ID, nil, err)
		return domain.Payment{}, err
	}

	if payment.SubscriptionID != current.SubscriptionID || payment.AmountCents != current.AmountCents || !sameDayTime(payment.PaidAt, current.PaidAt) {
		err := errors.New("para alterar valor ou data, estorne e registre um novo pagamento")
		recordAuditFailure(ctx, s.audit, "payment.update", "payment", payment.ID, nil, err)
		return domain.Payment{}, err
	}
	if payment.Status != current.Status {
		err := errors.New("use o estorno para alterar o status")
		recordAuditFailure(ctx, s.audit, "payment.update", "payment", payment.ID, nil, err)
		return domain.Payment{}, err
	}

	updated, err := s.repo.Update(ctx, payment)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "payment.update", "payment", payment.ID, nil, err)
		return domain.Payment{}, err
	}
	recordAuditSuccess(ctx, s.audit, "payment.update", "payment", updated.ID, map[string]any{
		"amount_cents":    updated.AmountCents,
		"credit_cents":    updated.CreditCents,
		"kind":            string(updated.Kind),
		"method":          string(updated.Method),
		"status":          string(updated.Status),
		"subscription_id": updated.SubscriptionID,
	})
	return updated, nil
}

func (s *PaymentService) FindByID(ctx context.Context, paymentID string) (domain.Payment, error) {
	return s.repo.FindByID(ctx, paymentID)
}

func (s *PaymentService) Reverse(ctx context.Context, paymentID string) (domain.Payment, error) {
	recordAuditAttempt(ctx, s.audit, "payment.reverse", "payment", paymentID, nil)

	payment, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "payment.reverse", "payment", paymentID, nil, err)
		return domain.Payment{}, err
	}
	if payment.Status == domain.PaymentReversed {
		recordAuditSuccess(ctx, s.audit, "payment.reverse", "payment", payment.ID, map[string]any{
			"status": string(payment.Status),
		})
		return payment, nil
	}

	if s.allocations != nil && s.periods != nil {
		if s.subscriptions == nil {
			err := errors.New("assinaturas indisponiveis")
			recordAuditFailure(ctx, s.audit, "payment.reverse", "payment", payment.ID, nil, err)
			return domain.Payment{}, err
		}
		subscription, err := s.subscriptions.FindByID(ctx, payment.SubscriptionID)
		if err != nil {
			recordAuditFailure(ctx, s.audit, "payment.reverse", "payment", payment.ID, nil, err)
			return domain.Payment{}, err
		}
		if err := s.rollbackPayment(ctx, payment, subscription); err != nil {
			recordAuditFailure(ctx, s.audit, "payment.reverse", "payment", payment.ID, nil, err)
			return domain.Payment{}, err
		}
	}

	if payment.CreditCents > 0 && s.balances != nil {
		if _, err := s.balances.Add(ctx, payment.SubscriptionID, -payment.CreditCents); err != nil {
			recordAuditFailure(ctx, s.audit, "payment.reverse", "payment", payment.ID, nil, err)
			return domain.Payment{}, err
		}
		payment.CreditCents = 0
	}

	payment.Status = domain.PaymentReversed
	payment.Kind = domain.PaymentFull
	updated, err := s.repo.Update(ctx, payment)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "payment.reverse", "payment", payment.ID, nil, err)
		return domain.Payment{}, err
	}
	recordAuditSuccess(ctx, s.audit, "payment.reverse", "payment", updated.ID, map[string]any{
		"status":          string(updated.Status),
		"subscription_id": updated.SubscriptionID,
	})
	return updated, nil
}

func (s *PaymentService) ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.Payment, error) {
	return s.repo.ListBySubscription(ctx, subscriptionID)
}

func (s *PaymentService) ListByPeriod(ctx context.Context, start, end time.Time) ([]domain.Payment, error) {
	return s.repo.ListByPeriod(ctx, start, end)
}

type paymentApplicationResult struct {
	Kind        domain.PaymentKind
	CreditCents int64
}

func (s *PaymentService) applyPayment(ctx context.Context, payment domain.Payment, subscription domain.Subscription, today time.Time) (paymentApplicationResult, error) {
	if s.periods == nil || s.allocations == nil {
		return paymentApplicationResult{}, errors.New("periodos de cobranca indisponiveis")
	}

	paymentDay, err := effectivePaymentDay(subscription)
	if err != nil {
		return paymentApplicationResult{}, err
	}
	if _, err := refreshPeriodStatusesBySubscription(ctx, s.periods, subscription.ID, paymentDay, today); err != nil {
		return paymentApplicationResult{}, err
	}
	periods, err := s.periods.ListOpenBySubscription(ctx, subscription.ID)
	if err != nil {
		return paymentApplicationResult{}, err
	}

	remaining := payment.AmountCents
	partial := false

	for _, period := range periods {
		if remaining == 0 {
			break
		}
		if period.AmountPaidCents >= period.AmountDueCents {
			continue
		}

		applied := minInt64(remaining, period.AmountDueCents-period.AmountPaidCents)
		if applied <= 0 {
			continue
		}

		period.AmountPaidCents += applied
		period.Status = resolvePeriodStatus(period, today, paymentDay)
		if period.AmountPaidCents < period.AmountDueCents {
			partial = true
		}

		updated, err := s.periods.Update(ctx, period)
		if err != nil {
			return paymentApplicationResult{}, err
		}

		if err := s.allocations.Create(ctx, domain.PaymentAllocation{
			PaymentID:       payment.ID,
			BillingPeriodID: updated.ID,
			AmountCents:     applied,
		}); err != nil {
			return paymentApplicationResult{}, err
		}

		remaining -= applied
	}

	credit := remaining
	if credit > 0 {
		if s.balances == nil {
			return paymentApplicationResult{}, errors.New("saldo indisponivel para registrar credito")
		}
		if _, err := s.balances.Add(ctx, subscription.ID, credit); err != nil {
			return paymentApplicationResult{}, err
		}
	}

	return paymentApplicationResult{
		Kind:        paymentKind(partial, false, credit > 0),
		CreditCents: credit,
	}, nil
}

func (s *PaymentService) rollbackPayment(ctx context.Context, payment domain.Payment, subscription domain.Subscription) error {
	if s.periods == nil || s.allocations == nil {
		return errors.New("dependencias de estorno indisponiveis")
	}

	allocations, err := s.allocations.ListByPayment(ctx, payment.ID)
	if err != nil {
		return err
	}

	if len(allocations) == 0 {
		return nil
	}

	periods, err := s.periods.ListBySubscription(ctx, payment.SubscriptionID)
	if err != nil {
		return err
	}

	periodMap := make(map[string]domain.BillingPeriod, len(periods))
	for _, period := range periods {
		periodMap[period.ID] = period
	}

	today := dateOnly(s.now())
	paymentDay, err := effectivePaymentDay(subscription)
	if err != nil {
		return err
	}
	for _, allocation := range allocations {
		period, ok := periodMap[allocation.BillingPeriodID]
		if !ok {
			continue
		}
		period.AmountPaidCents -= allocation.AmountCents
		if period.AmountPaidCents < 0 {
			period.AmountPaidCents = 0
		}
		period.Status = resolvePeriodStatus(period, today, paymentDay)
		if _, err := s.periods.Update(ctx, period); err != nil {
			return err
		}
	}

	return s.allocations.DeleteByPayment(ctx, payment.ID)
}

func (s *PaymentService) ensurePeriods(ctx context.Context, subscription domain.Subscription, plan domain.Plan, today time.Time) ([]domain.BillingPeriod, error) {
	return ensureBillingPeriods(ctx, s.periods, subscription, plan, today)
}

func (s *PaymentService) applyBalance(ctx context.Context, subscription domain.Subscription, today time.Time) error {
	return applySubscriptionBalance(ctx, s.balances, s.periods, subscription, today)
}

func buildPeriod(subscriptionID string, start time.Time, durationDays int, amountDue int64) domain.BillingPeriod {
	startDate := dateOnly(start)
	endDate := startDate.AddDate(0, 0, durationDays)

	return domain.BillingPeriod{
		SubscriptionID: subscriptionID,
		PeriodStart:    startDate,
		PeriodEnd:      endDate,
		AmountDueCents: amountDue,
		Status:         domain.BillingOpen,
	}
}

func resolvePeriodStatus(period domain.BillingPeriod, today time.Time, paymentDay int) domain.BillingPeriodStatus {
	dueDate := dueDateForPeriod(period.PeriodStart, paymentDay)
	if period.AmountPaidCents >= period.AmountDueCents {
		return domain.BillingPaid
	}
	if period.AmountPaidCents > 0 {
		if today.After(dueDate) {
			return domain.BillingOverdue
		}
		return domain.BillingPartial
	}
	if today.After(dueDate) {
		return domain.BillingOverdue
	}
	return domain.BillingOpen
}

func paymentKind(partial bool, paidFuture bool, hasCredit bool) domain.PaymentKind {
	switch {
	case hasCredit:
		return domain.PaymentCredit
	case partial:
		return domain.PaymentPartial
	case paidFuture:
		return domain.PaymentAdvance
	default:
		return domain.PaymentFull
	}
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func sameDayTime(a, b time.Time) bool {
	return a.Equal(b)
}
