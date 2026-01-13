package domain

import "time"

type Subscription struct {
	ID         string
	StudentID  string
	PlanID     string
	StartDate  time.Time
	EndDate    time.Time
	Status     SubscriptionStatus
	PriceCents int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
