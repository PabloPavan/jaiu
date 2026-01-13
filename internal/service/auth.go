package service

import (
	"context"
	"errors"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"golang.org/x/crypto/bcrypt"
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

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	if !user.Active {
		return domain.User{}, ports.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return domain.User{}, ports.ErrUnauthorized
		}
		return domain.User{}, err
	}

	return user, nil
}

func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}
