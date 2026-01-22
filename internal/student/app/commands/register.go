package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	p "github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/student/app/ports"
	"github.com/PabloPavan/jaiu/internal/student/domain"
)

type RegisterStudentRequest struct {
	FullName       string
	BirthDate      time.Time
	Gender         string
	Phone          string
	Email          string
	CPF            string
	Address        string
	Notes          string
	PhotoObjectKey string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type RegisterStudentResponse struct {
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
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type RegisterStudent struct {
	Repo  ports.StudentWriteRepository
	Audit p.AuditRepository
}

func NewRegisterStudent(repo ports.StudentWriteRepository, audit p.AuditRepository) *RegisterStudent {
	return &RegisterStudent{Repo: repo, Audit: audit}
}

func (rs *RegisterStudent) Execute(ctx context.Context, student RegisterStudentRequest) (RegisterStudentResponse, error) {
	var err error

	status, err := parseStatus(student.Status)
	if err != nil {
		return RegisterStudentResponse{}, err
	}

	s := domain.Student{
		FullName:       student.FullName,
		BirthDate:      student.BirthDate,
		Gender:         student.Gender,
		Phone:          student.Phone,
		Email:          student.Email,
		CPF:            student.CPF,
		Address:        student.Address,
		Status:         status,
		Notes:          student.Notes,
		PhotoObjectKey: student.PhotoObjectKey,
	}

	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now

	// metadata := map[string]any{
	// 	"status": string(student.Status),
	// }
	//rs.AuditRecordAuditAttempt(ctx, s.audit, "student.create", "student", student.ID, metadata)

	s, err = rs.Repo.Create(ctx, s)
	if err != nil {
		//	recordAuditFailure(ctx, s.audit, "student.create", "student", student.ID, metadata, err)
		return RegisterStudentResponse{}, err
	}
	//recordAuditSuccess(ctx, s.audit, "student.create", "student", created.ID, metadata)

	created := RegisterStudentResponse{
		FullName:       s.FullName,
		BirthDate:      s.BirthDate,
		Gender:         s.Gender,
		Phone:          s.Phone,
		Email:          s.Email,
		CPF:            s.CPF,
		Address:        s.Address,
		Status:         string(status),
		Notes:          s.Notes,
		PhotoObjectKey: s.PhotoObjectKey,
	}
	return created, nil
}

func parseStatus(value string) (domain.StudentStatus, error) {
	status := domain.StudentStatus(strings.ToLower(value))
	if status == "" {
		return domain.StudentActive, nil
	}
	if !status.IsValid() {
		return "", errors.New("Status invalido.")
	}
	return status, nil
}
