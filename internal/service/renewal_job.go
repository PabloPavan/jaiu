package service

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/ports"
)

type RenewalJob struct {
	subscriptions ports.SubscriptionRepository
	plans         ports.PlanRepository
	periods       ports.BillingPeriodRepository
	balances      ports.SubscriptionBalanceRepository
	now           func() time.Time
}

func NewRenewalJob(
	subscriptions ports.SubscriptionRepository,
	plans ports.PlanRepository,
	periods ports.BillingPeriodRepository,
	balances ports.SubscriptionBalanceRepository,
) *RenewalJob {
	return &RenewalJob{
		subscriptions: subscriptions,
		plans:         plans,
		periods:       periods,
		balances:      balances,
		now:           time.Now,
	}
}

func (j *RenewalJob) Run(ctx context.Context) error {
	if j.subscriptions == nil || j.plans == nil || j.periods == nil || j.balances == nil {
		return errors.New("dependencias de renovacao indisponiveis")
	}

	today := dateOnly(j.now())
	subscriptions, err := j.subscriptions.ListAutoRenew(ctx)
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		plan, err := j.plans.FindByID(ctx, subscription.PlanID)
		if err != nil {
			return err
		}
		if _, err := ensureBillingPeriods(ctx, j.periods, subscription, plan, today); err != nil {
			return err
		}
		if err := applySubscriptionBalance(ctx, j.balances, j.periods, subscription, today); err != nil {
			return err
		}
	}

	return nil
}
