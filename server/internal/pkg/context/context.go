package context

import (
	"context"
	"server/internal/domain"
)

type contextKey struct{}

func WithSession(ctx context.Context, session *domain.Session) context.Context {
	return context.WithValue(ctx, contextKey{}, session)
}

func SessionFromContext(ctx context.Context) (*domain.Session, bool) {
	session, ok := ctx.Value(contextKey{}).(*domain.Session)
	return session, ok
}

func MustSessionFromContext(ctx context.Context) *domain.Session {
	session, ok := SessionFromContext(ctx)
	if !ok {
		panic("session not found in context")
	}
	return session
}
