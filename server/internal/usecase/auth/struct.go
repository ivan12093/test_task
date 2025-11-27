package auth

import "log/slog"

type UseCase struct {
	logger       *slog.Logger
	userRepo     UserRepository
	sessionRepo  SessionRepository
	oauthGateway OAuthGateway
	csrfUC       CSRFTokenGenerator
}

func NewUseCase(logger *slog.Logger, userRepo UserRepository, sessionRepo SessionRepository, oauthGateway OAuthGateway, csrfUC CSRFTokenGenerator) *UseCase {
	return &UseCase{
		logger:       logger,
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		oauthGateway: oauthGateway,
		csrfUC:       csrfUC,
	}
}
