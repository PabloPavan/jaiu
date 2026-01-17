package service

import (
	"context"
	"errors"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo  ports.UserRepository
	audit ports.AuditRepository
}

func NewAuthService(repo ports.UserRepository, audit ports.AuditRepository) *AuthService {
	return &AuthService{repo: repo, audit: audit}
}

func (s *AuthService) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return s.repo.Create(ctx, user)
}

func (s *AuthService) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
	metadata := map[string]any{
		"email": email,
	}
	recordAuditAttempt(ctx, s.audit, "user.login", "user", "", metadata)

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "user.login", "user", "", metadata, err)
		return domain.User{}, err
	}

	if !user.Active {
		recordAuditFailure(ctx, s.audit, "user.login", "user", user.ID, metadata, ports.ErrUnauthorized)
		return domain.User{}, ports.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			recordAuditFailure(ctx, s.audit, "user.login", "user", user.ID, metadata, ports.ErrUnauthorized)
			return domain.User{}, ports.ErrUnauthorized
		}
		recordAuditFailure(ctx, s.audit, "user.login", "user", user.ID, metadata, err)
		return domain.User{}, err
	}

	recordAuditSuccess(ctx, s.audit, "user.login", "user", user.ID, metadata)
	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, session ports.Session) error {
	metadata := map[string]any{
		"user_id": session.UserID,
	}
	recordAuditAttempt(ctx, s.audit, "user.logout", "user", session.UserID, metadata)
	if session.UserID == "" {
		err := errors.New("missing session user")
		recordAuditFailure(ctx, s.audit, "user.logout", "user", "", metadata, err)
		return err
	}
	recordAuditSuccess(ctx, s.audit, "user.logout", "user", session.UserID, metadata)
	return nil
}

func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}
