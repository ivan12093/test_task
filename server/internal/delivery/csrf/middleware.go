package csrf

import (
	"errors"
	"log/slog"
	"net/http"
	"server/internal/pkg/httptools"
	"time"
)

const csrfTokenHeaderName = "X-CSRF-Token"
const csrfTokenCookieName = "csrf_token"
const csrfTokenCookieMaxAge = time.Hour * 24

type CSRFMiddleware struct {
	logger *slog.Logger
	uc     CSRFTokenUC
}

func NewCSRFMiddleware(logger *slog.Logger, uc CSRFTokenUC) *CSRFMiddleware {
	return &CSRFMiddleware{
		logger: logger,
		uc:     uc,
	}
}

func (m *CSRFMiddleware) SetCSRFToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := m.uc.GetCSRFToken(r.Context())
		if err != nil {
			m.logger.Error("failed to get CSRF token", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to get CSRF token")
			return
		}
		secure := r.TLS != nil
		http.SetCookie(w, &http.Cookie{
			Name:     csrfTokenCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: false,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(csrfTokenCookieMaxAge.Seconds()),
			Expires:  time.Now().Add(csrfTokenCookieMaxAge),
		})
		next.ServeHTTP(w, r)

	})
}

func (m *CSRFMiddleware) RequireCSRFToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(csrfTokenCookieName)
		if errors.Is(err, http.ErrNoCookie) {
			httptools.WriteJSONError(w, http.StatusBadRequest, "csrf token is required")
			return
		}
		if err != nil {
			m.logger.Error("failed to get CSRF token", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to get CSRF token")
			return
		}
		tokenHeader := r.Header.Get(csrfTokenHeaderName)
		ok, err := m.uc.ValidateCSRFToken(r.Context(), tokenCookie.Value, tokenHeader)
		if err != nil {
			m.logger.Error("failed to validate CSRF token", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to validate CSRF token")
			return
		}
		if !ok {
			httptools.WriteJSONError(w, http.StatusBadRequest, "invalid CSRF token")
			return
		}
		next.ServeHTTP(w, r)
	})
}
