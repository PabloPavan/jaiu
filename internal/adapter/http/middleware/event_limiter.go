package middleware

import (
	"context"
	"net/http"
	"sync"
)

// EventLimiter ensures only one active SSE connection per session cookie.
type EventLimiter struct {
	cookieName string
	mu         sync.Mutex
	cancels    map[string]*cancelEntry
}

func NewEventLimiter(cookieName string) *EventLimiter {
	return &EventLimiter{
		cookieName: cookieName,
		cancels:    make(map[string]*cancelEntry),
	}
}

func (l *EventLimiter) Wrap(next http.Handler) http.Handler {
	if l == nil || l.cookieName == "" || next == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(l.cookieName)
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		key := cookie.Value
		ctx, cancel := context.WithCancel(r.Context())
		entry := &cancelEntry{cancel: cancel}

		l.mu.Lock()
		if previous, ok := l.cancels[key]; ok && previous != nil && previous.cancel != nil {
			previous.cancel()
		}
		l.cancels[key] = entry
		l.mu.Unlock()

		defer func() {
			l.mu.Lock()
			if current, ok := l.cancels[key]; ok && current == entry {
				delete(l.cancels, key)
			}
			l.mu.Unlock()
			cancel()
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type cancelEntry struct {
	cancel context.CancelFunc
}
