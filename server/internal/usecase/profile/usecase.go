package profile

import (
	"context"
	"server/internal/domain"
)

func (uc *UseCase) GetProfile(ctx context.Context, userID int64) (*domain.Profile, error) {
	profile, err := uc.profileRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (uc *UseCase) UpdateProfile(ctx context.Context, userID int64, profile *domain.Profile) error {
	return uc.profileRepo.UpdateProfile(ctx, &domain.Profile{
		UserID:   userID,
		FullName: profile.FullName,
		Phone:    profile.Phone,
		Email:    profile.Email,
	})
}
