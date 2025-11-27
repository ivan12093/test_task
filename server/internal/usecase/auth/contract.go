package auth

import (
	"context"
	"server/internal/domain"
)

type UserRepository interface {
	CreateUserWithCredentials(ctx context.Context, credentials domain.Credentials) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByOAuthInfo(ctx context.Context, oauthInfo *domain.OAuthUserInfo) (*domain.User, error)
	CreateUserWithOAuthInfo(ctx context.Context, oauthInfo *domain.OAuthUserInfo) error
}

type SessionRepository interface {
	StoreSession(ctx context.Context, session *domain.Session) error
	GetSessionByToken(ctx context.Context, token string) (*domain.Session, error)
	DeleteSession(ctx context.Context, token string) error
}

type OAuthGateway interface {
	GetOAuthUserInfo(ctx context.Context, code, purpose string) (*domain.OAuthUserInfo, error)
	GetGoogleAuthURL(ctx context.Context, purpose, state string) string
}

type CSRFTokenGenerator interface {
	GetCSRFToken(ctx context.Context) (string, error)
}
