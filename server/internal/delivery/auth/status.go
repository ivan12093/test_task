package delivery

import (
	"errors"
	"net/http"

	"server/internal/domain"
	"server/internal/pkg/httptools"
)

func (h *Handler) CheckAuthStatus(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("auth_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			httptools.WriteJSONResponse(w, http.StatusOK, map[string]bool{"authenticated": false})
			return
		}
		h.logger.Error("failed to get auth token", "error", err)
		httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to get auth token")
		return
	}

	session, err := h.sessionUC.GetSessionByToken(r.Context(), token.Value)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			httptools.WriteJSONResponse(w, http.StatusOK, map[string]bool{"authenticated": false})
			return
		}
		h.logger.Error("failed to check session", "error", err)
		httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to check session")
		return
	}

	if session != nil {
		httptools.WriteJSONResponse(w, http.StatusOK, map[string]bool{"authenticated": true})
		return
	}

	httptools.WriteJSONResponse(w, http.StatusOK, map[string]bool{"authenticated": false})
}
