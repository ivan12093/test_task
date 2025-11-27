package delivery

import (
	"net/http"
	"server/internal/pkg/context"
	"server/internal/pkg/httptools"
	"time"
)

func (h *Handler) LogOut(w http.ResponseWriter, r *http.Request) {
	session := context.MustSessionFromContext(r.Context())
	err := h.uc.LogOut(r.Context(), session)
	if err != nil {
		h.logger.Error("failed to log out", "error", err)
		httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to log out")
		return
	}

	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	httptools.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}
