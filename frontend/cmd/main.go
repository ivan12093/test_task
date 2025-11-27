package main

import (
	"context"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"frontend/internal/config"
	authDelivery "frontend/internal/delivery/auth"
	profileDelivery "frontend/internal/delivery/profile"
	authGateway "frontend/internal/gateway/auth"
	profileGateway "frontend/internal/gateway/profile"
	"frontend/internal/pkg/ping"
)

func main() {
	configPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		logger.Error("failed to parse templates", "error", err)
		os.Exit(1)
	}

	authGW := authGateway.NewGateway(cfg.API.BaseURL)
	profileGW := profileGateway.NewGateway(cfg.API.BaseURL)

	pingClient := ping.NewClient(cfg.API.BaseURL)
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := pingClient.Ping(pingCtx); err != nil {
		logger.Error("backend is not available", "error", err, "api", cfg.API.BaseURL)
		os.Exit(1)
	} else {
		logger.Info("backend is available", "api", cfg.API.BaseURL)
	}
	pingCancel()

	authHandler := authDelivery.NewHandler(logger, templates, authGW)
	profileHandler := profileDelivery.NewHandler(logger, templates, profileGW)

	handler := SetupRoutes(RoutesConfig{
		AuthHandler:    authHandler,
		ProfileHandler: profileHandler,
		Templates:      templates,
	})

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
		logger.Info("starting frontend server", "addr", cfg.Server.Addr, "api", cfg.API.BaseURL)
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

	logger.Info("server exited gracefully")
}
