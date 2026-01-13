package service

import (
	"context"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type StudentService struct {
	repo ports.StudentRepository
	now  func() time.Time
}

func NewStudentService(repo ports.StudentRepository) *StudentService {
	return &StudentService{repo: repo, now: time.Now}
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
	return s.repo.Update(ctx, student)
}

func (s *StudentService) SetStatus(ctx context.Context, studentID string, status domain.StudentStatus) (domain.Student, error) {
	student, err := s.repo.FindByID(ctx, studentID)
	if err != nil {
		return domain.Student{}, err
	}

	student.Status = status
	student.UpdatedAt = s.now()

	return s.repo.Update(ctx, student)
}

func (s *StudentService) Search(ctx context.Context, filter ports.StudentFilter) ([]domain.Student, error) {
	return s.repo.Search(ctx, filter)
}
