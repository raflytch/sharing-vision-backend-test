package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"sharing-vision-backend-test/internal/config"
	"sharing-vision-backend-test/internal/database"
	articlehandler "sharing-vision-backend-test/internal/handler"
	"sharing-vision-backend-test/internal/repository"
	"sharing-vision-backend-test/internal/service"
	"sharing-vision-backend-test/internal/validator"
)

var application struct {
	sync.Once
	router  http.Handler
	initErr error
}

// Handler is the Vercel serverless entrypoint.
func Handler(response http.ResponseWriter, request *http.Request) {
	application.Do(initialize)
	if application.initErr != nil {
		writeInitializationError(response)
		return
	}

	// Vercel rewrites /article/... to /api/article/.... Gin owns /article/... routes.
	requestForGin := request
	if strings.HasPrefix(request.URL.Path, "/api") {
		requestForGin = request.Clone(request.Context())
		requestForGin.URL.Path = strings.TrimPrefix(request.URL.Path, "/api")
		if requestForGin.URL.Path == "" {
			requestForGin.URL.Path = "/"
		}
	}
	application.router.ServeHTTP(response, requestForGin)
}

func initialize() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	connectionContext, cancel := context.WithTimeout(context.Background(), cfg.QueryTimeout)
	defer cancel()

	db, err := database.OpenMySQL(connectionContext, cfg)
	if err != nil {
		application.initErr = err
		return
	}

	articleRepository := repository.NewMySQLArticleRepository(db, cfg.QueryTimeout)
	articleService := service.NewArticleService(articleRepository, validator.NewArticleValidator())
	articleHandler := articlehandler.NewArticleHandler(articleService, logger)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	articleHandler.RegisterRoutes(router)
	application.router = router
}

func writeInitializationError(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(response).Encode(map[string]string{
		"error": "service initialization failed",
	})
}
