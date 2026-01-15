package service

import (
	"context"
	"errors"
	"sort"
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
	now           func() time.Time
}

func NewPaymentService(
	repo ports.PaymentRepository,
	subscriptions ports.SubscriptionRepository,
	plans ports.PlanRepository,
	periods ports.BillingPeriodRepository,
	balances ports.SubscriptionBalanceRepository,
	allocations ports.PaymentAllocationRepository,
) *PaymentService {
	return &PaymentService{
		repo:          repo,
		subscriptions: subscriptions,
		plans:         plans,
		periods:       periods,
		balances:      balances,
		allocations:   allocations,
		now:           time.Now,
	}
}

func (s *PaymentService) Register(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	if payment.SubscriptionID == "" {
		return domain.Payment{}, errors.New("assinatura e obrigatoria")
	}
	if payment.AmountCents <= 0 {
		return domain.Payment{}, errors.New("valor deve ser maior que zero")
	}

	if s.subscriptions == nil || s.plans == nil || s.periods == nil || s.allocations == nil {
		return domain.Payment{}, errors.New("dependencias de pagamentos indisponiveis")
	}

	subscription, err := s.subscriptions.FindByID(ctx, payment.SubscriptionID)
	if err != nil {
		return domain.Payment{}, err
	}

	plan, err := s.plans.FindByID(ctx, subscription.PlanID)
	if err != nil {
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

	today := dateOnly(s.now())
	if _, err := s.ensurePeriods(ctx, subscription, plan, today); err != nil {
		return domain.Payment{}, err
	}
	if err := s.applyBalance(ctx, subscription, today); err != nil {
		return domain.Payment{}, err
	}

	created, err := s.repo.Create(ctx, payment)
	if err != nil {
		return domain.Payment{}, err
	}

	result, err := s.applyPayment(ctx, created, subscription, plan, today)
	if err != nil {
		return created, err
	}

	created.Kind = result.Kind
	created.CreditCents = result.CreditCents
	updated, err := s.repo.Update(ctx, created)
	if err != nil {
		return created, err
	}

	return updated, nil
}

func (s *PaymentService) Update(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	if payment.ID == "" {
		return domain.Payment{}, errors.New("pagamento invalido")
	}

	current, err := s.repo.FindByID(ctx, payment.ID)
	if err != nil {
		return domain.Payment{}, err
	}

	if payment.SubscriptionID == "" {
		payment.SubscriptionID = current.SubscriptionID
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

	if payment.SubscriptionID != current.SubscriptionID || payment.AmountCents != current.AmountCents || !sameDayTime(payment.PaidAt, current.PaidAt) {
		return domain.Payment{}, errors.New("para alterar valor ou data, estorne e registre um novo pagamento")
	}
	if payment.Status != current.Status {
		return domain.Payment{}, errors.New("use o estorno para alterar o status")
	}

	return s.repo.Update(ctx, payment)
}

func (s *PaymentService) FindByID(ctx context.Context, paymentID string) (domain.Payment, error) {
	return s.repo.FindByID(ctx, paymentID)
}

func (s *PaymentService) Reverse(ctx context.Context, paymentID string) (domain.Payment, error) {
	payment, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return domain.Payment{}, err
	}
	if payment.Status == domain.PaymentReversed {
		return payment, nil
	}

	if s.allocations != nil && s.periods != nil {
		if s.subscriptions == nil {
			return domain.Payment{}, errors.New("assinaturas indisponiveis")
		}
		subscription, err := s.subscriptions.FindByID(ctx, payment.SubscriptionID)
		if err != nil {
			return domain.Payment{}, err
		}
		if err := s.rollbackPayment(ctx, payment, subscription); err != nil {
			return domain.Payment{}, err
		}
	}

	if payment.CreditCents > 0 && s.balances != nil {
		if _, err := s.balances.Add(ctx, payment.SubscriptionID, -payment.CreditCents); err != nil {
			return domain.Payment{}, err
		}
		payment.CreditCents = 0
	}

	payment.Status = domain.PaymentReversed
	payment.Kind = domain.PaymentFull
	return s.repo.Update(ctx, payment)
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

func (s *PaymentService) applyPayment(ctx context.Context, payment domain.Payment, subscription domain.Subscription, plan domain.Plan, today time.Time) (paymentApplicationResult, error) {
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

	priceCents, err := effectivePriceCents(subscription, plan)
	if err != nil {
		return paymentApplicationResult{}, err
	}

	remaining := payment.AmountCents
	partial := false
	paidFuture := false
	lastStart := subscription.StartDate

	if allPeriods, err := s.periods.ListBySubscription(ctx, subscription.ID); err == nil && len(allPeriods) > 0 {
		sort.Slice(allPeriods, func(i, j int) bool {
			return allPeriods[i].PeriodStart.Before(allPeriods[j].PeriodStart)
		})
		lastStart = allPeriods[len(allPeriods)-1].PeriodStart
	}

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

	if remaining > 0 && subscription.AutoRenew {
		for remaining >= priceCents && priceCents > 0 {
			nextStart := renewalDateForPeriod(lastStart, paymentDay)
			future := buildPeriod(subscription.ID, nextStart, plan.DurationDays, priceCents)
			future.AmountPaidCents = priceCents
			future.Status = domain.BillingPaid

			created, err := s.periods.Create(ctx, future)
			if err != nil {
				return paymentApplicationResult{}, err
			}

			if err := s.allocations.Create(ctx, domain.PaymentAllocation{
				PaymentID:       payment.ID,
				BillingPeriodID: created.ID,
				AmountCents:     priceCents,
			}); err != nil {
				return paymentApplicationResult{}, err
			}

			remaining -= priceCents
			lastStart = created.PeriodStart
			paidFuture = true
		}
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
		Kind:        paymentKind(partial, paidFuture, credit > 0),
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
	if s.balances == nil || s.periods == nil {
		return nil
	}

	balance, err := s.balances.Get(ctx, subscription.ID)
	if err != nil {
		return err
	}

	remaining := balance.CreditCents
	if remaining <= 0 {
		return nil
	}

	paymentDay, err := effectivePaymentDay(subscription)
	if err != nil {
		return err
	}
	if _, err := refreshPeriodStatusesBySubscription(ctx, s.periods, subscription.ID, paymentDay, today); err != nil {
		return err
	}

	periods, err := s.periods.ListOpenBySubscription(ctx, subscription.ID)
	if err != nil {
		return err
	}

	for _, period := range periods {
		if remaining == 0 {
			break
		}
		due := period.AmountDueCents - period.AmountPaidCents
		if due <= 0 {
			continue
		}

		applied := minInt64(remaining, due)
		period.AmountPaidCents += applied
		period.Status = resolvePeriodStatus(period, today, paymentDay)
		if _, err := s.periods.Update(ctx, period); err != nil {
			return err
		}
		remaining -= applied
	}

	if remaining != balance.CreditCents {
		_, err = s.balances.Set(ctx, domain.SubscriptionBalance{
			SubscriptionID: subscription.ID,
			CreditCents:    remaining,
		})
		if err != nil {
			return err
		}
	}

	return nil
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
