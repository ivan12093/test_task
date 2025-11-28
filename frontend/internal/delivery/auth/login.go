package auth

import (
	"fmt"
	"net/http"
	"net/url"

	"frontend/internal/domain"
)

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		opts := pageDataOptions{}
		if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
			opts.Error = errorMsg
		} else if successMsg := r.URL.Query().Get("success"); successMsg != "" {
			opts.Message = successMsg
		}
		h.showLoginForm(w, r, newLoginPageData(opts))
	case http.MethodPost:
		h.handleLogin(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) showLoginForm(w http.ResponseWriter, r *http.Request, data pageData) {
	result, err := h.authGateway.CheckAuthStatus(r.Context())
	if err != nil || result.Status == domain.ResponseStatusError {
		h.logger.Error("failed to check auth status", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.logger.Info("auth status", "is_authenticated", result.IsAuthenticated)
	if result.IsAuthenticated {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	setCookies(w, result.Cookies)
	err = h.templates.ExecuteTemplate(w, "auth.html", data)
	if err != nil {
		h.logger.Error("failed to render login page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/login?error=%s", url.QueryEscape("Failed to process form")), http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Redirect(w, r, fmt.Sprintf("/login?error=%s", url.QueryEscape("Please fill in all fields")), http.StatusSeeOther)
		return
	}

	result, err := h.authGateway.Login(r.Context(), email, password)
	if err != nil {
		h.logger.Error("failed to login", "error", err)
		http.Redirect(w, r, fmt.Sprintf("/login?error=%s", url.QueryEscape("Failed to connect to server")), http.StatusSeeOther)
		return
	}

	if result.Status == domain.ResponseStatusError {
		http.Redirect(w, r, fmt.Sprintf("/login?error=%s", url.QueryEscape(result.Error)), http.StatusSeeOther)
		return
	}

	setCookies(w, result.Cookies)

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	result, err := h.authGateway.GetGoogleAuthURL(r.Context(), domain.GoogleAuthPurposeLogin)
	if err != nil {
		h.logger.Error("failed to get google auth URL", "error", err)
		http.Redirect(w, r, fmt.Sprintf("/login?error=%s", url.QueryEscape("Failed to get Google sign-in link")), http.StatusSeeOther)
		return
	}

	setCookies(w, result.Cookies)

	http.Redirect(w, r, result.URL, http.StatusTemporaryRedirect)
}
