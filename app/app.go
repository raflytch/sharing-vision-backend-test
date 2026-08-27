package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"sharing-vision-backend-test/internal/config"
	"sharing-vision-backend-test/internal/database"
	articlehandler "sharing-vision-backend-test/internal/handler"
	"sharing-vision-backend-test/internal/repository"
	"sharing-vision-backend-test/internal/service"
	"sharing-vision-backend-test/internal/validator"
)

// NewHandler builds the HTTP application and its database pool.
// The caller owns the returned database and should close it when appropriate.
func NewHandler(ctx context.Context, logger *slog.Logger) (http.Handler, *sql.DB, error) {
	cfg := config.Load()
	connectionContext, cancel := context.WithTimeout(ctx, cfg.QueryTimeout)
	defer cancel()

	db, err := database.OpenMySQL(connectionContext, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize application database: %w", err)
	}

	articleRepository := repository.NewMySQLArticleRepository(db, cfg.QueryTimeout)
	articleService := service.NewArticleService(articleRepository, validator.NewArticleValidator())
	articleHandler := articlehandler.NewArticleHandler(articleService, logger)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	articleHandler.RegisterRoutes(router)
	return router, db, nil
}

// NewLogger creates the structured logger shared by application adapters.
func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
