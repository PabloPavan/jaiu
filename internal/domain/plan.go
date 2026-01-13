package domain

import "time"

type Plan struct {
	ID           string
	Name         string
	DurationDays int
	PriceCents   int64
	Active       bool
	Description  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
