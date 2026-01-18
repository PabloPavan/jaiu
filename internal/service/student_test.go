package service

import (
	"context"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa Register validando nome e status.
func TestStudentServiceRegisterValidation(t *testing.T) {
	service := NewStudentService(&studentRepoFake{}, nil, nil)

	if _, err := service.Register(context.Background(), domain.Student{}); err == nil {
		t.Fatal("expected error for empty name")
	}
	if _, err := service.Register(context.Background(), domain.Student{FullName: "Name", Status: domain.StudentStatus("x")}); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

// Testa Register setando defaults e timestamps.
func TestStudentServiceRegisterDefaults(t *testing.T) {
	repo := &studentRepoFake{}
	service := NewStudentService(repo, nil, nil)
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	student, err := service.Register(context.Background(), domain.Student{FullName: " John "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if student.Status != domain.StudentActive {
		t.Fatalf("expected active status, got %q", student.Status)
	}
	if !student.CreatedAt.Equal(now) || !student.UpdatedAt.Equal(now) {
		t.Fatal("expected timestamps to be set")
	}
}

// Testa Update encerrando assinaturas quando aluno fica inativo.
func TestStudentServiceUpdateEndsSubscriptions(t *testing.T) {
	studentRepo := &studentRepoFake{}
	subRepo := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {ID: "sub-1", StudentID: "student-1", Status: domain.SubscriptionActive},
		},
	}
	service := NewStudentService(studentRepo, subRepo, nil)
	service.now = func() time.Time { return time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) }

	_, err := service.Update(context.Background(), domain.Student{ID: "student-1", FullName: "Name", Status: domain.StudentInactive})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subRepo.subscriptions["sub-1"].Status != domain.SubscriptionEnded {
		t.Fatalf("expected subscription to be ended, got %q", subRepo.subscriptions["sub-1"].Status)
	}
}

// Testa Deactivate setando status inativo.
func TestStudentServiceDeactivate(t *testing.T) {
	studentRepo := &studentRepoFake{
		students: map[string]domain.Student{
			"student-1": {ID: "student-1", FullName: "Name", Status: domain.StudentActive},
		},
	}
	subRepo := &subscriptionRepoFake{}
	service := NewStudentService(studentRepo, subRepo, nil)

	student, err := service.Deactivate(context.Background(), "student-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if student.Status != domain.StudentInactive {
		t.Fatalf("expected inactive, got %q", student.Status)
	}
}
