package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/marlendd/pt-start/internal/config"
	"github.com/marlendd/pt-start/internal/httpapi"
	"github.com/marlendd/pt-start/internal/shortener"
	"github.com/marlendd/pt-start/internal/storage/postgres"
)

func main() {
	if err := run(); err != nil {
		slog.Error(
			"application stopped",
			"error",
			err,
		)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := newLogger(cfg.LogLevel)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := openDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	logger.Info("connected to database")

	repository := postgres.NewRepository(pool)

	service := shortener.NewService(
		repository,
		shortener.RandomGenerator{},
	)

	handler := httpapi.NewHandler(
		service,
		pool,
		logger,
		cfg.BaseURL,
	)

	router := httpapi.NewRouter(handler)

	server := newHTTPServer(
		cfg.HTTPAddr,
		router,
		logger,
	)

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info(
			"starting HTTP server",
			"address",
			cfg.HTTPAddr,
		)

		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil

	case <-ctx.Done():
		stop()
		logger.Info("shutting down HTTP server")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if err := <-serverErrors; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}

	logger.Info("HTTP server stopped")

	return nil
}
