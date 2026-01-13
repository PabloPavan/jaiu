package ports

import (
	"context"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

type Session struct {
	UserID    string
	Name      string
	Role      domain.UserRole
	ExpiresAt time.Time
}

type SessionStore interface {
	Create(ctx context.Context, session Session) (string, error)
	Get(ctx context.Context, token string) (Session, error)
	Delete(ctx context.Context, token string) error
}
