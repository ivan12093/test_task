package profile

import (
	"context"
	"frontend/internal/domain"
)

type Gateway interface {
	GetProfile(ctx context.Context) (*domain.ProfileResult, error)
	UpdateProfile(ctx context.Context, profile *domain.Profile) (*domain.ProfileResult, error)
}
