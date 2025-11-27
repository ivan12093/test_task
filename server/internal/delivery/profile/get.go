package profile

import (
	"net/http"
	"server/internal/pkg/context"
	"server/internal/pkg/httptools"
)

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	session := context.MustSessionFromContext(r.Context())
	profile, err := h.uc.GetProfile(r.Context(), session.UserID)

	if err != nil {
		h.logger.Error("failed to get profile", "error", err)
		httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to get profile")
		return
	}

	dto := profileDTO{}
	dto.FromDomain(profile)

	httptools.WriteJSONResponse(w, http.StatusOK, dto)
}
