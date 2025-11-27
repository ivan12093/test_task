package profile

import (
	"fmt"
	"net/http"
	"net/url"

	"frontend/internal/domain"
)

func (h *Handler) ViewProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := h.profileGateway.GetProfile(r.Context())
	if err != nil {
		h.logger.Error("failed to get profile", "error", err)
		h.showProfileView(w, r, profileViewData{
			Error: "Failed to connect to server",
		})
		return
	}

	if result.Status == domain.ResponseStatusError {
		if result.StatusCode == http.StatusUnauthorized {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		h.showProfileView(w, r, profileViewData{
			Error: result.Error,
		})
		return
	}

	if result.Profile.FullName == "" {
		http.Redirect(w, r, "/profile/edit", http.StatusSeeOther)
		return
	}

	setCookies(w, result.Cookies)

	data := profileViewData{
		FullName: result.Profile.FullName,
		Phone:    result.Profile.Phone,
		Email:    result.Profile.Email,
	}

	if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
		data.Error = errorMsg
	}
	if successMsg := r.URL.Query().Get("success"); successMsg != "" {
		data.Success = successMsg
	}

	h.showProfileView(w, r, data)
}

func (h *Handler) EditProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		result, err := h.profileGateway.GetProfile(r.Context())
		if err != nil {
			h.logger.Error("failed to get profile", "error", err)
			h.showProfileEdit(w, r, profileEditData{
				Error: "Failed to connect to server",
			})
			return
		}

		if result.Status == domain.ResponseStatusError {
			if result.StatusCode == http.StatusUnauthorized {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			h.showProfileEdit(w, r, profileEditData{
				Error: result.Error,
			})
			return
		}

		setCookies(w, result.Cookies)

		data := profileEditData{
			FullName: result.Profile.FullName,
			Phone:    result.Profile.Phone,
			Email:    result.Profile.Email,
		}

		// Читаем query параметры для сообщений
		if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
			data.Error = errorMsg
		}
		if successMsg := r.URL.Query().Get("success"); successMsg != "" {
			data.Success = successMsg
		}

		h.showProfileEdit(w, r, data)
	case http.MethodPost:
		h.handleProfileUpdate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleProfileUpdate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/profile/edit?error=%s", url.QueryEscape("Failed to process form")), http.StatusSeeOther)
		return
	}

	profile := &domain.Profile{
		FullName: r.FormValue("full_name"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
	}

	result, err := h.profileGateway.UpdateProfile(r.Context(), profile)
	if err != nil {
		h.logger.Error("failed to update profile", "error", err)
		http.Redirect(w, r, fmt.Sprintf("/profile/edit?error=%s", url.QueryEscape("Failed to connect to server")), http.StatusSeeOther)
		return
	}

	setCookies(w, result.Cookies)

	if result.Status == domain.ResponseStatusError {
		http.Redirect(w, r, fmt.Sprintf("/profile/edit?error=%s", url.QueryEscape(result.Error)), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/profile?success=%s", url.QueryEscape("Profile updated successfully")), http.StatusSeeOther)
}

func setCookies(w http.ResponseWriter, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		http.SetCookie(w, cookie)
	}
}

func (h *Handler) showProfileView(w http.ResponseWriter, _ *http.Request, data profileViewData) {
	err := h.templates.ExecuteTemplate(w, "profile-view.html", data)
	if err != nil {
		h.logger.Error("failed to render profile view page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) showProfileEdit(w http.ResponseWriter, _ *http.Request, data profileEditData) {
	err := h.templates.ExecuteTemplate(w, "profile-edit.html", data)
	if err != nil {
		h.logger.Error("failed to render profile edit page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
