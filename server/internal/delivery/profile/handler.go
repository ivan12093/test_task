package profile

import (
	"log/slog"
)

type Handler struct {
	logger *slog.Logger
	uc     ProfileUC
}

func NewHandler(logger *slog.Logger, uc ProfileUC) *Handler {
	return &Handler{
		logger: logger,
		uc:     uc,
	}
}
