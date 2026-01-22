package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/PabloPavan/jaiu/internal/auditctx"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type contextKey string

const sessionKey contextKey = "session"

func RequireSession(store ports.SessionStore, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if store == nil {
				http.Error(w, "sessoes nao configuradas", http.StatusNotImplemented)
				return
			}

			cookie, err := r.Cookie(cookieName)
			if err != nil || cookie.Value == "" {
				deny(w, r)
				return
			}

			session, err := store.Get(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, ports.ErrNotFound) {
					deny(w, r)
					return
				}
				http.Error(w, "erro ao validar sessao", http.StatusInternalServerError)
				return
			}

			if activity, ok := userActivityFromContext(r.Context()); ok {
				activity.UserID = session.UserID
				activity.Role = string(session.Role)
			}

			ctx := context.WithValue(r.Context(), sessionKey, session)
			ctx = auditctx.WithActor(ctx, auditctx.Actor{
				ID:   session.UserID,
				Role: string(session.Role),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SessionFromContext(ctx context.Context) (ports.Session, bool) {
	session, ok := ctx.Value(sessionKey).(ports.Session)
	return session, ok
}

func deny(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	http.Error(w, "nao autorizado", http.StatusUnauthorized)
}
