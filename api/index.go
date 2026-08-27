package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"sharing-vision-backend-test/app"
)

var application struct {
	sync.Once
	router  http.Handler
	db      io.Closer
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
	logger := app.NewLogger()
	router, db, err := app.NewHandler(context.Background(), logger)
	if err != nil {
		application.initErr = err
		return
	}
	application.router = router
	application.db = db
}

func writeInitializationError(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(response).Encode(map[string]string{
		"error": "service initialization failed",
	})
}
