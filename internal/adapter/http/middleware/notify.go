package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PabloPavan/eventrail/sse"
	"github.com/go-chi/chi/v5"
)

type NotifyConfig struct {
	Publisher *sse.Publisher
	Scope     string
	ScopeID   int64
	Context   context.Context
}

const originHeader = "X-Origin-ID"

func Notify(cfg NotifyConfig) func(http.Handler) http.Handler {
	if cfg.Publisher == nil || strings.TrimSpace(cfg.Scope) == "" {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	scopeID := cfg.ScopeID
	if scopeID == 0 {
		scopeID = 1
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !shouldNotifyMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			recorder := &notifyRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r)

			if recorder.status >= http.StatusBadRequest {
				return
			}

			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = r.URL.Path
			}
			topic := topicFromRoute(route)
			if topic == "" {
				return
			}

			payload := map[string]any{
				"method": r.Method,
				"route":  route,
				"path":   r.URL.Path,
			}
			originID := strings.TrimSpace(r.Header.Get(originHeader))
			if originID == "" {
				originID = strings.TrimSpace(r.URL.Query().Get("origin_id"))
			}
			if originID == "" {
				originID = strings.TrimSpace(r.Form.Get("origin_id"))
			}
			if originID != "" {
				payload["origin_id"] = originID
			}
			data, err := json.Marshal(payload)
			if err != nil {
				return
			}

			baseCtx := cfg.Context
			if baseCtx == nil {
				baseCtx = r.Context()
			}
			ctx, cancel := context.WithTimeout(baseCtx, 2*time.Second)
			defer cancel()

			_ = cfg.Publisher.PublishEvent(ctx, fmt.Sprintf("%s:%d:%s", cfg.Scope, scopeID, topic), sse.Event{
				EventType: topic + ".changed",
				Data:      json.RawMessage(data),
			})
		})
	}
}

func shouldNotifyMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func topicFromRoute(route string) string {
	route = strings.TrimSpace(route)
	if route == "" {
		return ""
	}
	route = strings.Split(route, "?")[0]
	route = strings.TrimPrefix(route, "/")
	if route == "" {
		return "home"
	}

	parts := strings.Split(route, "/")
	topic := strings.ToLower(parts[0])
	switch topic {
	case "static", "images", "events", "healthz":
		return ""
	default:
		return topic
	}
}

type notifyRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *notifyRecorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *notifyRecorder) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(data)
}
