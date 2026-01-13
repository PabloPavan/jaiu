package service

import (
	"context"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type AuthService struct {
	repo ports.UserRepository
}

func NewAuthService(repo ports.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return s.repo.Create(ctx, user)
}

func (s *AuthService) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.repo.FindByEmail(ctx, email)
}
