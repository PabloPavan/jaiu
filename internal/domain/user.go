package domain

import "time"

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         UserRole
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
