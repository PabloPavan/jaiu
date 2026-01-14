package domain

import "time"

type PaymentAllocation struct {
	PaymentID       string
	BillingPeriodID string
	AmountCents     int64
	CreatedAt       time.Time
}
