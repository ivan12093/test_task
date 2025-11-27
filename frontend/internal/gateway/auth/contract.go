package auth

import (
	"context"
	"frontend/internal/domain"
)

type AuthGateway interface {
	Login(ctx context.Context, email, password string) (*domain.LoginResult, error)
	SignUp(ctx context.Context, email, password string) (*domain.SignUpResult, error)
	GetGoogleAuthURL(ctx context.Context, purpose domain.GoogleAuthPurpose) (*domain.GoogleAuthResult, error)
	Logout(ctx context.Context) error
}
