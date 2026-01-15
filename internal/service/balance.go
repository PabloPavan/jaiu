package service

import (
	"context"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

func applySubscriptionBalance(
	ctx context.Context,
	balances ports.SubscriptionBalanceRepository,
	periods ports.BillingPeriodRepository,
	subscription domain.Subscription,
	today time.Time,
) error {
	if balances == nil || periods == nil {
		return nil
	}

	balance, err := balances.Get(ctx, subscription.ID)
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
	if _, err := refreshPeriodStatusesBySubscription(ctx, periods, subscription.ID, paymentDay, today); err != nil {
		return err
	}

	openPeriods, err := periods.ListOpenBySubscription(ctx, subscription.ID)
	if err != nil {
		return err
	}

	for _, period := range openPeriods {
		if remaining == 0 {
			break
		}
		due := period.AmountDueCents - period.AmountPaidCents
		if due <= 0 {
			continue
		}

		applied := remaining
		if applied > due {
			applied = due
		}

		period.AmountPaidCents += applied
		period.Status = resolvePeriodStatus(period, today, paymentDay)
		if _, err := periods.Update(ctx, period); err != nil {
			return err
		}
		remaining -= applied
	}

	if remaining != balance.CreditCents {
		_, err = balances.Set(ctx, domain.SubscriptionBalance{
			SubscriptionID: subscription.ID,
			CreditCents:    remaining,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
