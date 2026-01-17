package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type StudentService struct {
	repo          ports.StudentRepository
	subscriptions ports.SubscriptionRepository
	audit         ports.AuditRepository
	now           func() time.Time
}

func NewStudentService(repo ports.StudentRepository, subscriptions ports.SubscriptionRepository, audit ports.AuditRepository) *StudentService {
	return &StudentService{repo: repo, subscriptions: subscriptions, audit: audit, now: time.Now}
}

func (s *StudentService) Register(ctx context.Context, student domain.Student) (domain.Student, error) {
	student.FullName = strings.TrimSpace(student.FullName)
	if student.FullName == "" {
		return domain.Student{}, errors.New("nome completo e obrigatorio")
	}
	if student.Status == "" {
		student.Status = domain.StudentActive
	}
	if !student.Status.IsValid() {
		return domain.Student{}, errors.New("status do aluno invalido")
	}

	now := s.now()
	student.CreatedAt = now
	student.UpdatedAt = now

	metadata := map[string]any{
		"status": string(student.Status),
	}
	recordAuditAttempt(ctx, s.audit, "student.create", "student", student.ID, metadata)

	created, err := s.repo.Create(ctx, student)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "student.create", "student", student.ID, metadata, err)
		return domain.Student{}, err
	}
	recordAuditSuccess(ctx, s.audit, "student.create", "student", created.ID, metadata)
	return created, nil
}

func (s *StudentService) Update(ctx context.Context, student domain.Student) (domain.Student, error) {
	if student.ID == "" {
		return domain.Student{}, errors.New("aluno invalido")
	}
	student.FullName = strings.TrimSpace(student.FullName)
	if student.FullName == "" {
		return domain.Student{}, errors.New("nome completo e obrigatorio")
	}
	if student.Status == "" {
		student.Status = domain.StudentActive
	}
	if !student.Status.IsValid() {
		return domain.Student{}, errors.New("status do aluno invalido")
	}
	student.UpdatedAt = s.now()
	metadata := map[string]any{
		"status": string(student.Status),
	}
	recordAuditAttempt(ctx, s.audit, "student.update", "student", student.ID, metadata)

	updated, err := s.repo.Update(ctx, student)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "student.update", "student", student.ID, metadata, err)
		return domain.Student{}, err
	}
	if updated.Status == domain.StudentInactive || updated.Status == domain.StudentSuspended {
		if err := s.endSubscriptionsForStudent(ctx, updated.ID); err != nil {
			recordAuditFailure(ctx, s.audit, "student.update", "student", updated.ID, metadata, err)
			return updated, err
		}
	}
	recordAuditSuccess(ctx, s.audit, "student.update", "student", updated.ID, metadata)
	return updated, nil
}

func (s *StudentService) SetStatus(ctx context.Context, studentID string, status domain.StudentStatus) (domain.Student, error) {
	student, err := s.repo.FindByID(ctx, studentID)
	if err != nil {
		return domain.Student{}, err
	}

	student.Status = status
	return s.Update(ctx, student)
}

func (s *StudentService) FindByID(ctx context.Context, studentID string) (domain.Student, error) {
	return s.repo.FindByID(ctx, studentID)
}

func (s *StudentService) Deactivate(ctx context.Context, studentID string) (domain.Student, error) {
	recordAuditAttempt(ctx, s.audit, "student.deactivate", "student", studentID, nil)
	updated, err := s.SetStatus(ctx, studentID, domain.StudentInactive)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "student.deactivate", "student", studentID, nil, err)
		return domain.Student{}, err
	}
	recordAuditSuccess(ctx, s.audit, "student.deactivate", "student", updated.ID, map[string]any{
		"status": string(updated.Status),
	})
	return updated, nil
}

func (s *StudentService) Search(ctx context.Context, filter ports.StudentFilter) ([]domain.Student, error) {
	return s.repo.Search(ctx, filter)
}

func (s *StudentService) endSubscriptionsForStudent(ctx context.Context, studentID string) error {
	if s.subscriptions == nil {
		return errors.New("assinaturas indisponiveis")
	}
	subscriptions, err := s.subscriptions.ListByStudent(ctx, studentID)
	if err != nil {
		return err
	}
	return endSubscriptions(ctx, s.subscriptions, subscriptions, s.now())
}
