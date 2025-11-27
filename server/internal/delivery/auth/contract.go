package delivery

import (
	"context"
	"server/internal/domain"
)

type SessionUC interface {
	GetSessionByToken(ctx context.Context, token string) (*domain.Session, error)
}

type AuthUC interface {
	SignUpWithEmail(ctx context.Context, email, password string) error
	LogInWithEmail(ctx context.Context, email, password string) (*domain.Session, error)
	LogOut(ctx context.Context, session *domain.Session) error
	LogInWithGoogle(ctx context.Context, code string) (*domain.Session, error)
	SignUpWithGoogle(ctx context.Context, code string) error
	GetGoogleAuthURL(ctx context.Context, purpose string) (string, string, error)
}
