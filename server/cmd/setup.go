package main

import (
	"io"
	"net/http"

	authDelivery "server/internal/delivery/auth"
	csrfDelivery "server/internal/delivery/csrf"
	profileDelivery "server/internal/delivery/profile"
	middleware "server/internal/pkg/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type RoutesConfig struct {
	AuthHandler     *authDelivery.Handler
	ProfileHandler  *profileDelivery.Handler
	AuthMiddleware  *authDelivery.AuthMiddleware
	CSRFMiddleware  *csrfDelivery.CSRFMiddleware
	PanicMiddleware *middleware.PanicMiddleware
	CORSMiddleware  *cors.Cors
}

func SetupRoutes(config RoutesConfig) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(NotFound)
	router.Use(config.PanicMiddleware.PanicMiddleware)

	var corsRouter *mux.Router
	if config.CORSMiddleware != nil {
		corsRouter = router.Methods(http.MethodGet, http.MethodPost,
			http.MethodPut, http.MethodDelete, http.MethodOptions).Subrouter()
		corsRouter.Use(config.CORSMiddleware.Handler)
	} else {
		corsRouter = router
	}

	authRouter := corsRouter.Methods(http.MethodGet, http.MethodPost,
		http.MethodPut, http.MethodDelete, http.MethodOptions).Subrouter()
	authRouter.Use(config.AuthMiddleware.RequireAuth)

	unAuthRouter := corsRouter.Methods(http.MethodGet, http.MethodPost,
		http.MethodPut, http.MethodDelete, http.MethodOptions).Subrouter()
	unAuthRouter.Use(config.AuthMiddleware.RequireUnAuth, config.CSRFMiddleware.RequireCSRFToken, config.CSRFMiddleware.SetCSRFToken)

	unAuthRouter.HandleFunc("/api/auth/signup", config.AuthHandler.SignUpWithEmail).Methods(http.MethodPost)
	unAuthRouter.HandleFunc("/api/auth/login", config.AuthHandler.LogInWithEmail).Methods(http.MethodPost)
	unAuthRouter.HandleFunc("/api/auth/google/url", config.AuthHandler.GetGoogleAuthURL).Methods(http.MethodGet)

	if config.CORSMiddleware != nil {
		corsRouter.Handle("/api/auth/status", config.CSRFMiddleware.SetCSRFToken(http.HandlerFunc(config.AuthHandler.CheckAuthStatus))).Methods(http.MethodGet)
	} else {
		router.Handle("/api/auth/status", config.CSRFMiddleware.SetCSRFToken(http.HandlerFunc(config.AuthHandler.CheckAuthStatus))).Methods(http.MethodGet)
	}

	router.HandleFunc("/ping", Ping).Methods(http.MethodGet)

	authRouter.Handle("/api/auth/logout", config.CSRFMiddleware.SetCSRFToken(http.HandlerFunc(config.AuthHandler.LogOut))).Methods(http.MethodPost)
	authRouter.HandleFunc("/api/profile", config.ProfileHandler.GetProfile).Methods(http.MethodGet)
	authRouter.Handle("/api/profile", config.CSRFMiddleware.RequireCSRFToken(http.HandlerFunc(config.ProfileHandler.UpdateProfile))).Methods(http.MethodPut)

	unCorsedUnAuthRouter := router.Methods(http.MethodGet, http.MethodPost).Subrouter()
	unCorsedUnAuthRouter.Use(config.CSRFMiddleware.SetCSRFToken, config.AuthMiddleware.RequireUnAuth)

	unCorsedUnAuthRouter.HandleFunc("/api/auth/google/callback/login", config.AuthHandler.LogInWithGoogle).Methods(http.MethodGet)
	unCorsedUnAuthRouter.HandleFunc("/api/auth/google/callback/signup", config.AuthHandler.SignUpWithGoogle).Methods(http.MethodGet)

	return router
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"message": "not found"}`)
}

func Ping(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"message": "pong"}`)
}
