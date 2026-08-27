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

	"github.com/gin-gonic/gin"

	"sharing-vision-backend-test/internal/config"
	"sharing-vision-backend-test/internal/database"
	"sharing-vision-backend-test/internal/handler"
	"sharing-vision-backend-test/internal/repository"
	"sharing-vision-backend-test/internal/service"
	"sharing-vision-backend-test/internal/validator"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()

	connectionContext, cancelConnection := context.WithTimeout(context.Background(), cfg.QueryTimeout)
	defer cancelConnection()

	db, err := database.OpenMySQL(connectionContext, cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	articleRepository := repository.NewMySQLArticleRepository(db, cfg.QueryTimeout)
	articleValidator := validator.NewArticleValidator()
	articleService := service.NewArticleService(articleRepository, articleValidator)
	articleHandler := handler.NewArticleHandler(articleService, logger)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	articleHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.QueryTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("article service started", "address", cfg.ServerAddress)
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case serverErr := <-serverErrors:
		if !errors.Is(serverErr, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", serverErr)
		}
		return nil
	case <-shutdownSignal.Done():
		logger.Info("shutting down article service")
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	return nil
}
