package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/ports"
	redis "github.com/redis/go-redis/v9"
)

type SessionStore struct {
	client *redis.Client
	prefix string
}

func NewSessionStore(client *redis.Client) *SessionStore {
	return &SessionStore{client: client, prefix: "session:"}
}

func (s *SessionStore) Create(ctx context.Context, session ports.Session) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = time.Now().Add(24 * time.Hour)
	}

	payload, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	if err := s.client.Set(ctx, s.prefix+token, payload, ttl).Err(); err != nil {
		return "", err
	}

	return token, nil
}

func (s *SessionStore) Get(ctx context.Context, token string) (ports.Session, error) {
	if token == "" {
		return ports.Session{}, ports.ErrNotFound
	}

	value, err := s.client.Get(ctx, s.prefix+token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ports.Session{}, ports.ErrNotFound
		}
		return ports.Session{}, err
	}

	var session ports.Session
	if err := json.Unmarshal([]byte(value), &session); err != nil {
		return ports.Session{}, err
	}

	if !session.ExpiresAt.IsZero() && time.Now().After(session.ExpiresAt) {
		_ = s.client.Del(ctx, s.prefix+token).Err()
		return ports.Session{}, ports.ErrNotFound
	}

	return session, nil
}

func (s *SessionStore) Delete(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	return s.client.Del(ctx, s.prefix+token).Err()
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
