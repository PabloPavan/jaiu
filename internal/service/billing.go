package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

func ensureBillingPeriods(ctx context.Context, repo ports.BillingPeriodRepository, subscription domain.Subscription, plan domain.Plan, today time.Time) ([]domain.BillingPeriod, error) {
	if subscription.StartDate.IsZero() {
		return nil, errors.New("data de inicio da assinatura invalida")
	}
	if plan.DurationDays <= 0 {
		return nil, errors.New("duracao do plano invalida")
	}

	priceCents, err := effectivePriceCents(subscription, plan)
	if err != nil {
		return nil, err
	}

	paymentDay, err := effectivePaymentDay(subscription)
	if err != nil {
		return nil, err
	}

	periods, err := repo.ListBySubscription(ctx, subscription.ID)
	if err != nil {
		return nil, err
	}

	if len(periods) == 0 {
		first := buildPeriod(subscription.ID, subscription.StartDate, plan.DurationDays, priceCents)
		first.Status = resolvePeriodStatus(first, today, paymentDay)
		created, err := repo.Create(ctx, first)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				periods, err = repo.ListBySubscription(ctx, subscription.ID)
				if err != nil {
					return nil, err
				}
				if len(periods) == 0 {
					return nil, errors.New("periodo de cobranca indisponivel")
				}
			} else {
				return nil, err
			}
		} else {
			periods = append(periods, created)
		}
	}

	sort.Slice(periods, func(i, j int) bool {
		return periods[i].PeriodStart.Before(periods[j].PeriodStart)
	})

	if subscription.AutoRenew {
		periods, err = ensureRenewals(ctx, repo, periods, subscription, plan.DurationDays, priceCents, today, paymentDay)
		if err != nil {
			return nil, err
		}
	}

	return refreshPeriodStatuses(ctx, repo, periods, paymentDay, today)
}

func refreshPeriodStatuses(ctx context.Context, repo ports.BillingPeriodRepository, periods []domain.BillingPeriod, paymentDay int, today time.Time) ([]domain.BillingPeriod, error) {
	updated := make([]domain.BillingPeriod, 0, len(periods))
	for _, period := range periods {
		status := resolvePeriodStatus(period, today, paymentDay)
		if status != period.Status {
			period.Status = status
			saved, err := repo.Update(ctx, period)
			if err != nil {
				return nil, err
			}
			period = saved
		}
		updated = append(updated, period)
	}
	return updated, nil
}

func refreshPeriodStatusesBySubscription(ctx context.Context, repo ports.BillingPeriodRepository, subscriptionID string, paymentDay int, today time.Time) ([]domain.BillingPeriod, error) {
	periods, err := repo.ListBySubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	return refreshPeriodStatuses(ctx, repo, periods, paymentDay, today)
}

func ensureRenewals(
	ctx context.Context,
	repo ports.BillingPeriodRepository,
	periods []domain.BillingPeriod,
	subscription domain.Subscription,
	durationDays int,
	priceCents int64,
	today time.Time,
	paymentDay int,
) ([]domain.BillingPeriod, error) {
	if len(periods) == 0 {
		return periods, nil
	}

	periodByStart := make(map[string]domain.BillingPeriod, len(periods))
	for _, period := range periods {
		periodByStart[dateKey(period.PeriodStart)] = period
	}

	sort.Slice(periods, func(i, j int) bool {
		return periods[i].PeriodStart.Before(periods[j].PeriodStart)
	})

	last := periods[len(periods)-1]
	nextStart := renewalDateForPeriod(last.PeriodStart, paymentDay)
	for !nextStart.After(today) {
		key := dateKey(nextStart)
		if existing, ok := periodByStart[key]; ok {
			last = existing
		} else {
			future := buildPeriod(subscription.ID, nextStart, durationDays, priceCents)
			future.Status = resolvePeriodStatus(future, today, paymentDay)
			created, err := repo.Create(ctx, future)
			if err != nil {
				if errors.Is(err, ports.ErrConflict) {
					refreshed, err := repo.ListBySubscription(ctx, subscription.ID)
					if err != nil {
						return periods, err
					}
					periods = refreshed
					periodByStart = make(map[string]domain.BillingPeriod, len(periods))
					for _, period := range periods {
						periodByStart[dateKey(period.PeriodStart)] = period
					}
					existing, ok := periodByStart[key]
					if !ok {
						return periods, errors.New("periodo de cobranca indisponivel")
					}
					last = existing
					nextStart = renewalDateForPeriod(last.PeriodStart, paymentDay)
					continue
				}
				return periods, err
			}
			periods = append(periods, created)
			periodByStart[key] = created
			last = created
		}
		nextStart = renewalDateForPeriod(last.PeriodStart, paymentDay)
	}

	sort.Slice(periods, func(i, j int) bool {
		return periods[i].PeriodStart.Before(periods[j].PeriodStart)
	})

	return periods, nil
}

func effectivePriceCents(subscription domain.Subscription, plan domain.Plan) (int64, error) {
	priceCents := subscription.PriceCents
	if priceCents <= 0 {
		if plan.PriceCents <= 0 {
			return 0, errors.New("valor do plano invalido")
		}
		priceCents = plan.PriceCents
	}
	return priceCents, nil
}

func effectivePaymentDay(subscription domain.Subscription) (int, error) {
	if subscription.PaymentDay <= 0 {
		if subscription.StartDate.IsZero() {
			return 0, errors.New("dia do pagamento invalido")
		}
		return subscription.StartDate.Day(), nil
	}
	if subscription.PaymentDay < 1 || subscription.PaymentDay > 31 {
		return 0, errors.New("dia do pagamento invalido")
	}
	return subscription.PaymentDay, nil
}

func renewalDateForPeriod(start time.Time, paymentDay int) time.Time {
	return dueDateForPeriod(start, paymentDay).AddDate(0, 0, 1)
}

func dueDateForPeriod(start time.Time, paymentDay int) time.Time {
	startDate := dateOnly(start)
	year, month, _ := startDate.Date()
	loc := startDate.Location()

	day := clampPaymentDay(paymentDay, year, month, loc)
	due := time.Date(year, month, day, 0, 0, 0, 0, loc)
	if due.Before(startDate) {
		next := startDate.AddDate(0, 1, 0)
		year, month, _ = next.Date()
		day = clampPaymentDay(paymentDay, year, month, loc)
		due = time.Date(year, month, day, 0, 0, 0, 0, loc)
	}
	return due
}

func clampPaymentDay(paymentDay int, year int, month time.Month, loc *time.Location) int {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
	if paymentDay > lastDay {
		return lastDay
	}
	return paymentDay
}

func dateKey(value time.Time) string {
	return dateOnly(value).Format("2006-01-02")
}
