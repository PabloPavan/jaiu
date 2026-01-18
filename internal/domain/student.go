package domain

import "time"

type Student struct {
	ID             string
	FullName       string
	BirthDate      *time.Time
	Gender         string
	Phone          string
	Email          string
	CPF            string
	Address        string
	Notes          string
	PhotoObjectKey string
	Status         StudentStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
