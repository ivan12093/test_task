package auth

import (
	"html/template"
	"log/slog"
)

type Handler struct {
	logger      *slog.Logger
	templates   *template.Template
	authGateway AuthGateway
}

func NewHandler(logger *slog.Logger, templates *template.Template, authGateway AuthGateway) *Handler {
	return &Handler{
		logger:      logger,
		templates:   templates,
		authGateway: authGateway,
	}
}
