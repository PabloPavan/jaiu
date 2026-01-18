package service

import (
	"context"
	"testing"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

// Testa HashPassword gerando um hash valido para comparacao.
func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("secret")); err != nil {
		t.Fatalf("expected hash to match, got error: %v", err)
	}
}

// Testa Authenticate propagando erro do repositorio.
func TestAuthServiceAuthenticateRepoError(t *testing.T) {
	repo := &userRepoFake{findErr: context.Canceled}
	service := NewAuthService(repo, nil)

	if _, err := service.Authenticate(context.Background(), "user@example.com", "secret"); err == nil {
		t.Fatal("expected error from repository")
	}
}

// Testa Authenticate bloqueando usuario inativo.
func TestAuthServiceAuthenticateInactive(t *testing.T) {
	repo := &userRepoFake{
		users: map[string]domain.User{
			"user@example.com": {Email: "user@example.com", Active: false},
		},
	}
	service := NewAuthService(repo, nil)

	if _, err := service.Authenticate(context.Background(), "user@example.com", "secret"); err != ports.ErrUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

// Testa Authenticate bloqueando senha incorreta.
func TestAuthServiceAuthenticateWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	repo := &userRepoFake{
		users: map[string]domain.User{
			"user@example.com": {Email: "user@example.com", Active: true, PasswordHash: hash},
		},
	}
	service := NewAuthService(repo, nil)

	if _, err := service.Authenticate(context.Background(), "user@example.com", "wrong"); err != ports.ErrUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

// Testa Authenticate retornando usuario em caso de sucesso.
func TestAuthServiceAuthenticateSuccess(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := domain.User{ID: "user-1", Email: "user@example.com", Active: true, PasswordHash: hash}
	repo := &userRepoFake{
		users: map[string]domain.User{
			"user@example.com": expected,
		},
	}
	service := NewAuthService(repo, nil)

	user, err := service.Authenticate(context.Background(), "user@example.com", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != expected.ID {
		t.Fatalf("expected user %q, got %q", expected.ID, user.ID)
	}
}

// Testa Logout exigindo usuario na sessao.
func TestAuthServiceLogout(t *testing.T) {
	service := NewAuthService(&userRepoFake{}, nil)

	if err := service.Logout(context.Background(), ports.Session{}); err == nil {
		t.Fatal("expected error for empty session")
	}
	if err := service.Logout(context.Background(), ports.Session{UserID: "user-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
