package domain

import "time"

type Payment struct {
	ID             string
	SubscriptionID string
	PaidAt         time.Time
	AmountCents    int64
	Method         PaymentMethod
	Reference      string
	Notes          string
	Status         PaymentStatus
	CreatedAt      time.Time
}
