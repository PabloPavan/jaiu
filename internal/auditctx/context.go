package auditctx

import "context"

type Actor struct {
	ID    string
	Role  string
	Email string
}

type RequestInfo struct {
	IP        string
	UserAgent string
}

type Context struct {
	Actor   Actor
	Request RequestInfo
}

type contextKey string

const auditKey contextKey = "audit_context"

func WithRequest(ctx context.Context, info RequestInfo) context.Context {
	current := FromContext(ctx)
	current.Request = info
	return context.WithValue(ctx, auditKey, current)
}

func WithActor(ctx context.Context, actor Actor) context.Context {
	current := FromContext(ctx)
	current.Actor = actor
	return context.WithValue(ctx, auditKey, current)
}

func FromContext(ctx context.Context) Context {
	if ctx == nil {
		return Context{}
	}
	if value, ok := ctx.Value(auditKey).(Context); ok {
		return value
	}
	return Context{}
}
