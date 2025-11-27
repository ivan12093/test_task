package auth

import (
	"net/http"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := h.authGateway.Logout(r.Context())
	if err != nil {
		if result.StatusCode == http.StatusForbidden {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		h.logger.Error("failed to logout", "error", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	if result != nil {
		setCookies(w, result.Cookies)
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
