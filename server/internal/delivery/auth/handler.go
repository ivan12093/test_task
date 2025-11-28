package delivery

import (
	"log/slog"
	"server/internal/config"
)

type Handler struct {
	uc          AuthUC
	sessionUC   SessionUC
	logger      *slog.Logger
	frontendURL string
	config      *config.Config
}

func NewHandler(uc AuthUC, sessionUC SessionUC, logger *slog.Logger, frontendURL string, config *config.Config) *Handler {
	return &Handler{
		uc:          uc,
		sessionUC:   sessionUC,
		logger:      logger,
		frontendURL: frontendURL,
		config:      config,
	}
}
