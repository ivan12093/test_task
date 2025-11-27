package profile

import (
	"encoding/json"
	"net/http"
	"server/internal/pkg/context"
	"server/internal/pkg/httptools"
)

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	session := context.MustSessionFromContext(r.Context())
	profileDTO := profileDTO{}
	err := json.NewDecoder(r.Body).Decode(&profileDTO)
	if err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		httptools.WriteJSONError(w, http.StatusBadRequest, "bad input data")
		return
	}
	err = h.uc.UpdateProfile(r.Context(), session.UserID, profileDTO.ToDomain())
	if err != nil {
		h.logger.Error("failed to update profile", "error", err)
		httptools.WriteJSONError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}
	httptools.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "profile updated successfully"})
}
