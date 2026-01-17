package ports

import "context"

type PaymentDependencies struct {
	Payments       PaymentRepository
	Subscriptions  SubscriptionRepository
	Plans          PlanRepository
	BillingPeriods BillingPeriodRepository
	Balances       SubscriptionBalanceRepository
	Allocations    PaymentAllocationRepository
	Audit          AuditRepository
}

type PaymentTxRunner interface {
	RunSerializable(ctx context.Context, fn func(context.Context, PaymentDependencies) error) error
}
