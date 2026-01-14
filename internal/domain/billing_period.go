package domain

import "time"

type BillingPeriod struct {
	ID              string
	SubscriptionID  string
	PeriodStart     time.Time
	PeriodEnd       time.Time
	AmountDueCents  int64
	AmountPaidCents int64
	Status          BillingPeriodStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
