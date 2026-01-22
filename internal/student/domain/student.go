package domain

import "time"

type StudentStatus string

const (
	StudentActive    StudentStatus = "active"
	StudentInactive  StudentStatus = "inactive"
	StudentSuspended StudentStatus = "suspended"
)

type Student struct {
	ID             string
	FullName       string
	BirthDate      time.Time
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

func (s StudentStatus) IsValid() bool {
	switch s {
	case StudentActive, StudentInactive, StudentSuspended:
		return true
	default:
		return false
	}
}
