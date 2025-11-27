package delivery

import (
	"errors"
	"log/slog"
	"net/http"
	"server/internal/domain"
	"server/internal/pkg/context"
	"server/internal/pkg/httptools"
)

type AuthMiddleware struct {
	logger *slog.Logger
	uc     SessionUC
}

func NewAuthMiddleware(logger *slog.Logger, uc SessionUC) *AuthMiddleware {
	return &AuthMiddleware{
		logger: logger,
		uc:     uc,
	}
}

func (m *AuthMiddleware) RequireUnAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("auth_token")
		if errors.Is(err, http.ErrNoCookie) {
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			m.logger.Error("failed to get auth token", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to get auth token")
			return
		}

		_, err = m.uc.GetSessionByToken(r.Context(), token.Value)
		if errors.Is(err, domain.ErrSessionNotFound) {
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			m.logger.Error("failed to check if session is active", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to check if session is active")
			return
		}

		httptools.WriteJSONError(w, http.StatusForbidden, "session is active")
	})
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("auth_token")
		if errors.Is(err, http.ErrNoCookie) {
			httptools.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if err != nil {
			m.logger.Error("failed to get auth token", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to get auth token")
			return
		}
		session, err := m.uc.GetSessionByToken(r.Context(), token.Value)
		if errors.Is(err, domain.ErrSessionNotFound) {
			httptools.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if err != nil {
			m.logger.Error("failed to check if session is active", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to check if session is active")
			return
		}
		r = r.WithContext(context.WithSession(r.Context(), session))
		next.ServeHTTP(w, r)
	})
}
