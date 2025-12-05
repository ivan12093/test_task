package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"server/internal/config"
	authDelivery "server/internal/delivery/auth"
	csrfDelivery "server/internal/delivery/csrf"
	profileDelivery "server/internal/delivery/profile"
	authGateway "server/internal/gateway/google"
	middleware "server/internal/pkg/middleware"
	sessionRepo "server/internal/repository/session"
	userRepo "server/internal/repository/user"
	authUC "server/internal/usecase/auth"
	csrfUC "server/internal/usecase/csrf"
	profileUC "server/internal/usecase/profile"

	"github.com/rs/cors"
)

func main() {
	configPath := flag.String("config", "config.yml", "path to config file")
	envPath := flag.String("env", ".env", "path to .env file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load(*configPath, *envPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", cfg.Database.DSN())
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	userRepository := userRepo.NewRepository(logger, db)
	sessionRepository := sessionRepo.NewRepository()

	googleOAuthGateway := authGateway.NewOAuthGateway(authGateway.GoogleOAuthConfig{
		ClientID:     cfg.OAuth.Google.ClientID,
		ClientSecret: cfg.OAuth.Google.ClientSecret,
		RedirectURL:  cfg.OAuth.Google.RedirectURL,
	})

	csrfUseCase := csrfUC.NewUseCase(logger)
	profileUseCase := profileUC.NewUseCase(logger, userRepository)
	authUseCase := authUC.NewUseCase(logger, userRepository, sessionRepository, googleOAuthGateway, csrfUseCase)

	authHandler := authDelivery.NewHandler(authUseCase, sessionRepository, logger, cfg.Server.FrontendURL, cfg)
	profileHandler := profileDelivery.NewHandler(logger, profileUseCase)

	authMiddleware := authDelivery.NewAuthMiddleware(logger, sessionRepository)
	csrfMiddleware := csrfDelivery.NewCSRFMiddleware(logger, csrfUseCase)
	panicMiddleware := middleware.NewPanicMiddleware(logger)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)

	var corsMiddleware *cors.Cors
	if cfg.Server.CORSEnabled {
		corsMiddleware = cors.New(cors.Options{
			AllowedOrigins:   []string{cfg.Server.FrontendURL},
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowCredentials: true,
		})
		logger.Info("CORS enabled", "frontend_url", cfg.Server.FrontendURL)
	} else {
		logger.Info("CORS disabled")
	}

	router := SetupRoutes(RoutesConfig{
		AuthHandler:     authHandler,
		ProfileHandler:  profileHandler,
		AuthMiddleware:  authMiddleware,
		CSRFMiddleware:  csrfMiddleware,
		PanicMiddleware: panicMiddleware,
		CORSMiddleware:  corsMiddleware,
	})

	handler := loggingMiddleware.AccessLog(router)

	server := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		logger.Info("starting server", "addr", cfg.Server.Addr, "frontend", cfg.Server.FrontendURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	if err := db.Close(); err != nil {
		logger.Error("failed to close database", "error", err)
	}

	logger.Info("server exited gracefully")
}
