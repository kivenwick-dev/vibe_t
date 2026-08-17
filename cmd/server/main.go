package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"banner-fingerprint/internal/api"
	"banner-fingerprint/internal/fingerprint"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	rulesPath := getenv("RULES_PATH", "rules/rules.json")
	addr := getenv("LISTEN_ADDR", ":8080")

	engine, err := fingerprint.Load(rulesPath)
	if err != nil {
		logger.Error("load fingerprint rules", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           api.NewServer(engine, logger).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	logger.Info("server started", "addr", addr, "rules", rulesPath)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
