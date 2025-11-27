package profile

import "log/slog"

type UseCase struct {
	logger      *slog.Logger
	profileRepo ProfileRepository
}

func NewUseCase(logger *slog.Logger, profileRepo ProfileRepository) *UseCase {
	return &UseCase{
		logger:      logger,
		profileRepo: profileRepo,
	}
}
