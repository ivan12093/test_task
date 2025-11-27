package delivery

import "log/slog"

type Handler struct {
	uc          AuthUC
	sessionUC   SessionUC
	logger      *slog.Logger
	frontendURL string
}

func NewHandler(uc AuthUC, sessionUC SessionUC, logger *slog.Logger, frontendURL string) *Handler {
	return &Handler{
		uc:          uc,
		sessionUC:   sessionUC,
		logger:      logger,
		frontendURL: frontendURL,
	}
}
