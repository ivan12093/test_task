package main

import (
	"html/template"
	"net/http"

	authDelivery "frontend/internal/delivery/auth"
	profileDelivery "frontend/internal/delivery/profile"
	"frontend/internal/pkg/cache"
	"frontend/internal/pkg/cookies"
	"frontend/internal/pkg/logging"
	"frontend/internal/pkg/proxy"
)

type RoutesConfig struct {
	AuthHandler       *authDelivery.Handler
	ProfileHandler    *profileDelivery.Handler
	Templates         *template.Template
	LoggingMiddleware *logging.LoggingMiddleware
	ProxyHandler      *proxy.ProxyHandler
}

func SetupRoutes(config RoutesConfig) http.Handler {
	mux := http.NewServeMux()

	staticFS := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))

	if config.ProxyHandler != nil {
		mux.Handle("/api/", config.ProxyHandler)
	}

	mux.HandleFunc("/login", config.AuthHandler.LoginPage)
	mux.HandleFunc("/login/google", config.AuthHandler.GoogleLogin)

	mux.HandleFunc("/signup", config.AuthHandler.SignUpPage)
	mux.HandleFunc("/signup/google", config.AuthHandler.GoogleSignUp)

	mux.HandleFunc("/logout", config.AuthHandler.Logout)

	mux.HandleFunc("/profile", config.ProfileHandler.ViewProfile)
	mux.HandleFunc("/profile/edit", config.ProfileHandler.EditProfile)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			notFoundHandler(w, r, config.Templates)
			return
		}
		redirectURL := "/login"
		if success := r.URL.Query().Get("success"); success == "true" {
			redirectURL = "/login?success=true"
		}
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	})

	handler := config.LoggingMiddleware.AccessLog(mux)
	handler = cookies.Middleware(handler)
	handler = cache.NoCacheMiddleware(handler)
	return handler
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request, templates *template.Template) {
	w.WriteHeader(http.StatusNotFound)
	if err := templates.ExecuteTemplate(w, "404.html", nil); err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}
