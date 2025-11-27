package delivery

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"server/internal/domain"
	"server/internal/pkg/httptools"
	"time"
)

const stateCookieName = "state"
const stateCookieMaxAge = time.Minute * 10

func (h *Handler) GetGoogleAuthURL(w http.ResponseWriter, r *http.Request) {
	purpose := r.URL.Query().Get("purpose")
	if purpose != "login" && purpose != "signup" {
		httptools.WriteJSONError(w, http.StatusBadRequest, "invalid purpose")
		return
	}
	url, state, err := h.uc.GetGoogleAuthURL(r.Context(), purpose)
	if err != nil {
		h.logger.Error("failed to get google auth url", "error", err)
		httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to get google auth url")
		return
	}

	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(stateCookieMaxAge.Seconds()),
		Expires:  time.Now().Add(stateCookieMaxAge),
	})
	httptools.WriteJSONResponse(w, http.StatusOK, map[string]string{"url": url})
}

func (h *Handler) validateStateAndExtractCode(w http.ResponseWriter, r *http.Request, redirectURL *url.URL) (string, error) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		redirectURL.RawQuery = url.Values{"error": {"code parameter is required"}}.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return "", fmt.Errorf("code parameter is required")
	}

	if state == "" {
		redirectURL.RawQuery = url.Values{"error": {"state parameter is required"}}.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return "", fmt.Errorf("state parameter is required")
	}

	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil {
		redirectURL.RawQuery = url.Values{"error": {"state cookie is required"}}.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return "", fmt.Errorf("failed to get state cookie: %w", err)
	}
	if stateCookie.Value != state {
		redirectURL.RawQuery = url.Values{"error": {"state cookie is invalid"}}.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return "", fmt.Errorf("state cookie is invalid")
	}

	return code, nil
}

func (h *Handler) SignUpWithGoogle(w http.ResponseWriter, r *http.Request) {
	redirectURL, err := url.Parse(h.frontendURL + "/signup")
	if err != nil {
		h.logger.Error("internal server error", "error", err)
		redirectURL.RawQuery = url.Values{"error": {"internal server error"}}.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return
	}

	code, err := h.validateStateAndExtractCode(w, r, redirectURL)
	if err != nil {
		h.logger.Warn("failed to validate state and extract code", "error", err)
		return
	}

	err = h.uc.SignUpWithGoogle(r.Context(), code)
	if err != nil {
		var errorMessage string
		if errors.Is(err, domain.ErrInvalidGoogleCode) {
			errorMessage = "invalid google code"
		} else if errors.Is(err, domain.ErrUserAlreadyExists) {
			errorMessage = "user already exists"
		} else {
			errorMessage = "failed to sign up with google"
		}
		h.logger.Error("failed to sign up with google", "error", err)
		redirectURL.RawQuery = url.Values{"error": {errorMessage}}.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return
	}

	redirectURL.RawQuery = url.Values{"success": {"User created successfully"}}.Encode()
	http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
}

func (h *Handler) LogInWithGoogle(w http.ResponseWriter, r *http.Request) {
	redirectURL, err := url.Parse(h.frontendURL + "/login")
	if err != nil {
		h.logger.Error("internal server error", "error", err)
		redirectURL.RawQuery = url.Values{"error": {"internal server error"}}.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return
	}

	code, err := h.validateStateAndExtractCode(w, r, redirectURL)
	if err != nil {
		h.logger.Warn("failed to validate state and extract code", "error", err)
		return
	}

	session, err := h.uc.LogInWithGoogle(r.Context(), code)
	if err != nil {
		var errorMessage string
		if errors.Is(err, domain.ErrInvalidGoogleCode) {
			errorMessage = "invalid google code"
		} else if errors.Is(err, domain.ErrUserNotExists) {
			errorMessage = "user not exists"
		} else {
			errorMessage = "failed to log in with google"
		}
		h.logger.Error("failed to log in with google", "error", err)

		redirectURL, err := url.Parse(h.frontendURL + "/login")
		if err != nil {
			h.logger.Error("failed to parse redirect URL", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		redirectURL.RawQuery = url.Values{"error": {errorMessage}}.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return
	}

	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	redirectURL, err = url.Parse(h.frontendURL + "/profile")
	if err != nil {
		h.logger.Error("failed to parse redirect URL", "error", err)
		httptools.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
}
