package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"server/internal/domain"
	"server/internal/pkg/httptools"
	"time"
)

func (h *Handler) SignUpWithEmail(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userCreate := authDTO{}
	err := decoder.Decode(&userCreate)
	if err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		httptools.WriteJSONError(w, http.StatusBadRequest, "bad input data")
		return
	}

	err = h.uc.SignUpWithEmail(r.Context(), userCreate.Email, userCreate.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserAlreadyExists):
			h.logger.Info("user already exists", "email", userCreate.Email)
			httptools.WriteJSONError(w, http.StatusBadRequest, "user already exists")
		case errors.Is(err, domain.ErrNotValidEmail):
			h.logger.Warn("invalid email", "email", userCreate.Email)
			httptools.WriteJSONError(w, http.StatusBadRequest, "not valid email")
		default:
			h.logger.Error("internal error during sign up", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.logger.Info("user created successfully", "email", userCreate.Email)
	httptools.WriteJSONResponse(w, http.StatusCreated, map[string]string{"message": "user created successfully"})
}

func (h *Handler) LogInWithEmail(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userLogin := authDTO{}
	err := decoder.Decode(&userLogin)
	if err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		httptools.WriteJSONError(w, http.StatusBadRequest, "bad input data")
		return
	}

	session, err := h.uc.LogInWithEmail(r.Context(), userLogin.Email, userLogin.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidPassword):
			h.logger.Warn("invalid password", "email", userLogin.Email)
			httptools.WriteJSONError(w, http.StatusUnauthorized, "password is incorrect")
		case errors.Is(err, domain.ErrNotValidEmail):
			h.logger.Warn("invalid email", "email", userLogin.Email)
			httptools.WriteJSONError(w, http.StatusBadRequest, "not valid email")
		case errors.Is(err, domain.ErrUserNotExists):
			h.logger.Warn("user not exists", "email", userLogin.Email)
			httptools.WriteJSONError(w, http.StatusUnauthorized, "username entered does not exist")
		default:
			h.logger.Error("internal error during login", "error", err)
			httptools.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		}
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

	h.logger.Info("user logged in successfully", "email", userLogin.Email)
	httptools.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "logged in successfully"})
}
