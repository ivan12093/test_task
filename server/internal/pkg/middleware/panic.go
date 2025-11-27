package middleware

import (
	"log/slog"
	"net/http"
	httptools "server/internal/pkg/httptools"
)

type PanicMiddleware struct {
	logger *slog.Logger
}

func NewPanicMiddleware(logger *slog.Logger) *PanicMiddleware {
	return &PanicMiddleware{
		logger: logger,
	}
}

func (m *PanicMiddleware) PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				m.logger.Error("panic middleware", "error", r)
				httptools.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
