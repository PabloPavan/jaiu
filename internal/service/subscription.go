package service

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type SubscriptionService struct {
	repo     ports.SubscriptionRepository
	plans    ports.PlanRepository
	students ports.StudentRepository
	audit    ports.AuditRepository
	now      func() time.Time
}

func NewSubscriptionService(repo ports.SubscriptionRepository, plans ports.PlanRepository, students ports.StudentRepository, audit ports.AuditRepository) *SubscriptionService {
	return &SubscriptionService{
		repo:     repo,
		plans:    plans,
		students: students,
		audit:    audit,
		now:      time.Now,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	metadata := map[string]any{
		"auto_renew":  subscription.AutoRenew,
		"payment_day": subscription.PaymentDay,
		"plan_id":     subscription.PlanID,
		"price_cents": subscription.PriceCents,
		"status":      string(subscription.Status),
		"student_id":  subscription.StudentID,
	}
	recordAuditAttempt(ctx, s.audit, "subscription.create", "subscription", subscription.ID, metadata)

	if subscription.StudentID == "" {
		err := errors.New("aluno e obrigatorio")
		recordAuditFailure(ctx, s.audit, "subscription.create", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}
	if subscription.PlanID == "" {
		err := errors.New("plano e obrigatorio")
		recordAuditFailure(ctx, s.audit, "subscription.create", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}

	if s.students != nil {
		if _, err := s.students.FindByID(ctx, subscription.StudentID); err != nil {
			recordAuditFailure(ctx, s.audit, "subscription.create", "subscription", subscription.ID, metadata, err)
			return domain.Subscription{}, err
		}
	}

	var plan domain.Plan
	if s.plans != nil {
		loaded, err := s.plans.FindByID(ctx, subscription.PlanID)
		if err != nil {
			recordAuditFailure(ctx, s.audit, "subscription.create", "subscription", subscription.ID, metadata, err)
			return domain.Subscription{}, err
		}
		plan = loaded
	}

	if subscription.Status == "" {
		subscription.Status = domain.SubscriptionActive
	}

	if subscription.StartDate.IsZero() {
		subscription.StartDate = dateOnly(s.now())
	}

	if subscription.PaymentDay <= 0 {
		subscription.PaymentDay = subscription.StartDate.Day()
	}
	if subscription.PaymentDay < 1 || subscription.PaymentDay > 31 {
		err := errors.New("dia do pagamento invalido")
		recordAuditFailure(ctx, s.audit, "subscription.create", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}

	if subscription.EndDate.IsZero() && plan.DurationDays > 0 {
		subscription.EndDate = subscription.StartDate.AddDate(0, 0, plan.DurationDays)
	}

	if subscription.EndDate.IsZero() {
		err := errors.New("data de vencimento e obrigatoria")
		recordAuditFailure(ctx, s.audit, "subscription.create", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}

	if subscription.EndDate.Before(subscription.StartDate) {
		err := errors.New("data de vencimento deve ser depois da data de inicio")
		recordAuditFailure(ctx, s.audit, "subscription.create", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}

	if subscription.PriceCents <= 0 && plan.PriceCents > 0 {
		subscription.PriceCents = plan.PriceCents
	}

	now := s.now()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	created, err := s.repo.Create(ctx, subscription)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "subscription.create", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}
	recordAuditSuccess(ctx, s.audit, "subscription.create", "subscription", created.ID, metadata)
	return created, nil
}

func (s *SubscriptionService) Update(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	metadata := map[string]any{
		"auto_renew":  subscription.AutoRenew,
		"payment_day": subscription.PaymentDay,
		"plan_id":     subscription.PlanID,
		"price_cents": subscription.PriceCents,
		"status":      string(subscription.Status),
		"student_id":  subscription.StudentID,
	}
	recordAuditAttempt(ctx, s.audit, "subscription.update", "subscription", subscription.ID, metadata)

	if subscription.StartDate.IsZero() {
		err := errors.New("data de inicio e obrigatoria")
		recordAuditFailure(ctx, s.audit, "subscription.update", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}
	if subscription.Status == "" {
		subscription.Status = domain.SubscriptionActive
	}

	if subscription.PaymentDay < 1 || subscription.PaymentDay > 31 {
		err := errors.New("dia do pagamento invalido")
		recordAuditFailure(ctx, s.audit, "subscription.update", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}

	var plan domain.Plan
	if subscription.EndDate.IsZero() || subscription.PriceCents <= 0 {
		if s.plans == nil {
			err := errors.New("repositorio de planos nao configurado")
			recordAuditFailure(ctx, s.audit, "subscription.update", "subscription", subscription.ID, metadata, err)
			return domain.Subscription{}, err
		}
		loaded, err := s.plans.FindByID(ctx, subscription.PlanID)
		if err != nil {
			recordAuditFailure(ctx, s.audit, "subscription.update", "subscription", subscription.ID, metadata, err)
			return domain.Subscription{}, err
		}
		plan = loaded
	}

	if subscription.EndDate.IsZero() && plan.DurationDays > 0 {
		subscription.EndDate = subscription.StartDate.AddDate(0, 0, plan.DurationDays)
	}

	if subscription.EndDate.IsZero() {
		err := errors.New("data de vencimento e obrigatoria")
		recordAuditFailure(ctx, s.audit, "subscription.update", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}

	if subscription.EndDate.Before(subscription.StartDate) {
		err := errors.New("data de vencimento deve ser depois da data de inicio")
		recordAuditFailure(ctx, s.audit, "subscription.update", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}

	if subscription.PriceCents <= 0 && plan.PriceCents > 0 {
		subscription.PriceCents = plan.PriceCents
	}

	subscription.UpdatedAt = s.now()
	updated, err := s.repo.Update(ctx, subscription)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "subscription.update", "subscription", subscription.ID, metadata, err)
		return domain.Subscription{}, err
	}
	recordAuditSuccess(ctx, s.audit, "subscription.update", "subscription", updated.ID, metadata)
	return updated, nil
}

func (s *SubscriptionService) FindByID(ctx context.Context, subscriptionID string) (domain.Subscription, error) {
	return s.repo.FindByID(ctx, subscriptionID)
}

func (s *SubscriptionService) Cancel(ctx context.Context, subscriptionID string) (domain.Subscription, error) {
	recordAuditAttempt(ctx, s.audit, "subscription.cancel", "subscription", subscriptionID, nil)

	subscription, err := s.repo.FindByID(ctx, subscriptionID)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "subscription.cancel", "subscription", subscriptionID, nil, err)
		return domain.Subscription{}, err
	}

	subscription.Status = domain.SubscriptionCanceled
	subscription.UpdatedAt = s.now()

	updated, err := s.repo.Update(ctx, subscription)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "subscription.cancel", "subscription", subscriptionID, nil, err)
		return domain.Subscription{}, err
	}
	recordAuditSuccess(ctx, s.audit, "subscription.cancel", "subscription", updated.ID, map[string]any{
		"status": string(updated.Status),
	})
	return updated, nil
}

func (s *SubscriptionService) ListByStudent(ctx context.Context, studentID string) ([]domain.Subscription, error) {
	return s.repo.ListByStudent(ctx, studentID)
}

func (s *SubscriptionService) DueBetween(ctx context.Context, start, end time.Time) ([]domain.Subscription, error) {
	return s.repo.ListDueBetween(ctx, start, end)
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
