package service

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type RenewalJob struct {
	subscriptions ports.SubscriptionRepository
	plans         ports.PlanRepository
	periods       ports.BillingPeriodRepository
	balances      ports.SubscriptionBalanceRepository
	txRunner      ports.PaymentTxRunner
	now           func() time.Time
}

func NewRenewalJob(
	subscriptions ports.SubscriptionRepository,
	plans ports.PlanRepository,
	periods ports.BillingPeriodRepository,
	balances ports.SubscriptionBalanceRepository,
	txRunner ports.PaymentTxRunner,
) *RenewalJob {
	return &RenewalJob{
		subscriptions: subscriptions,
		plans:         plans,
		periods:       periods,
		balances:      balances,
		txRunner:      txRunner,
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
		if err := j.runSubscription(ctx, subscription, today); err != nil {
			return err
		}
	}

	return nil
}

func (j *RenewalJob) runSubscription(ctx context.Context, subscription domain.Subscription, today time.Time) error {
	if j.txRunner != nil {
		return j.txRunner.RunSerializable(ctx, func(ctx context.Context, deps ports.PaymentDependencies) error {
			return j.processSubscription(ctx, subscription, today, deps.Plans, deps.BillingPeriods, deps.Balances)
		})
	}

	return j.processSubscription(ctx, subscription, today, j.plans, j.periods, j.balances)
}

func (j *RenewalJob) processSubscription(
	ctx context.Context,
	subscription domain.Subscription,
	today time.Time,
	plans ports.PlanRepository,
	periods ports.BillingPeriodRepository,
	balances ports.SubscriptionBalanceRepository,
) error {
	plan, err := plans.FindByID(ctx, subscription.PlanID)
	if err != nil {
		return err
	}
	if _, err := ensureBillingPeriods(ctx, periods, subscription, plan, today); err != nil {
		return err
	}
	if err := applySubscriptionBalance(ctx, balances, periods, subscription, today); err != nil {
		return err
	}
	return nil
}
