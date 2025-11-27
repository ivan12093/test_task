package profile

import (
	"context"
	"server/internal/domain"
)

type ProfileUC interface {
	GetProfile(ctx context.Context, userID int64) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, userID int64, profile *domain.Profile) error
}
