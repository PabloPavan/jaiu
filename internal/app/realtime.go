package app

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"

	"github.com/PabloPavan/eventrail/sse"
	ssememory "github.com/PabloPavan/eventrail/sse/memory"
	sseredis "github.com/PabloPavan/eventrail/sse/redis"
	httpmw "github.com/PabloPavan/jaiu/internal/http/middleware"
	"github.com/PabloPavan/jaiu/internal/ports"
	redis "github.com/redis/go-redis/v9"
)

const (
	eventScope      = "app"
	eventNamePrefix = "app"
	eventScopeID    = int64(1)
)

type sessionResolver struct {
	store      ports.SessionStore
	cookieName string
}

func (r sessionResolver) Resolve(req *http.Request) (*sse.Principal, error) {
	if r.store == nil {
		return nil, errors.New("session store not configured")
	}

	cookie, err := req.Cookie(r.cookieName)
	if err != nil || cookie.Value == "" {
		return nil, errors.New("missing session cookie")
	}

	session, err := r.store.Get(req.Context(), cookie.Value)
	if err != nil {
		return nil, err
	}

	return &sse.Principal{
		UserID:  userIDToInt64(session.UserID),
		ScopeID: eventScopeID,
	}, nil
}

func newEventServer(ctx context.Context, redisClient *redis.Client, sessionStore ports.SessionStore, cookieName string) (*sse.Server, httpmw.NotifyConfig, error) {
	if sessionStore == nil {
		return nil, httpmw.NotifyConfig{}, nil
	}
	if ctx == nil {
		return nil, httpmw.NotifyConfig{}, errors.New("app context is required")
	}
	if cookieName == "" {
		cookieName = "jaiu_session"
	}

	var broker sse.Broker
	if redisClient != nil {
		broker = sseredis.NewBrokerPubSub(redisClient)
	} else {
		broker = ssememory.NewBrokerInMemory()
	}

	server, err := sse.NewServer(broker, sse.Options{
		Context: ctx,
		Resolver: sessionResolver{
			store:      sessionStore,
			cookieName: cookieName,
		},
		Router: func(p *sse.Principal) []string {
			return []string{fmt.Sprintf("%s:%d:*", eventScope, eventScopeID)}
		},
		EventNamePrefix: eventNamePrefix,
	})
	if err != nil {
		return nil, httpmw.NotifyConfig{}, err
	}

	notifyCfg := httpmw.NotifyConfig{
		Publisher: server.Publisher(),
		Scope:     eventScope,
		ScopeID:   eventScopeID,
		Context:   ctx,
	}

	return server, notifyCfg, nil
}

func userIDToInt64(value string) int64 {
	if value == "" {
		return 0
	}
	if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
		return parsed
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(value))
	return int64(hasher.Sum64())
}
