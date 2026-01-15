package service

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type StudentService struct {
	repo          ports.StudentRepository
	subscriptions ports.SubscriptionRepository
	now           func() time.Time
}

func NewStudentService(repo ports.StudentRepository, subscriptions ports.SubscriptionRepository) *StudentService {
	return &StudentService{repo: repo, subscriptions: subscriptions, now: time.Now}
}

func (s *StudentService) Register(ctx context.Context, student domain.Student) (domain.Student, error) {
	if student.Status == "" {
		student.Status = domain.StudentActive
	}

	now := s.now()
	student.CreatedAt = now
	student.UpdatedAt = now

	return s.repo.Create(ctx, student)
}

func (s *StudentService) Update(ctx context.Context, student domain.Student) (domain.Student, error) {
	student.UpdatedAt = s.now()
	updated, err := s.repo.Update(ctx, student)
	if err != nil {
		return domain.Student{}, err
	}
	if updated.Status == domain.StudentInactive || updated.Status == domain.StudentSuspended {
		if err := s.endSubscriptionsForStudent(ctx, updated.ID); err != nil {
			return updated, err
		}
	}
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
	return s.SetStatus(ctx, studentID, domain.StudentInactive)
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
