package domain

import "time"

type SubscriptionBalance struct {
	SubscriptionID string
	CreditCents    int64
	UpdatedAt      time.Time
}
