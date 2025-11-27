package profile

import (
	"frontend/internal/gateway/profile"
	"html/template"
	"log/slog"
)

type Handler struct {
	logger         *slog.Logger
	templates      *template.Template
	profileGateway profile.Gateway
}

func NewHandler(logger *slog.Logger, templates *template.Template, profileGateway profile.Gateway) *Handler {
	return &Handler{
		logger:         logger,
		templates:      templates,
		profileGateway: profileGateway,
	}
}
