package profile

import (
	"context"
	"server/internal/domain"
)

type ProfileRepository interface {
	GetProfileByUserID(ctx context.Context, userID int64) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, profile *domain.Profile) error
}
