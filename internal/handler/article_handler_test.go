package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"sharing-vision-backend-test/internal/model"
	"sharing-vision-backend-test/internal/service"
	"sharing-vision-backend-test/internal/validator"
)

type serviceStub struct {
	createFunc func(context.Context, model.ArticleInput) (model.Article, error)
	listFunc   func(context.Context, int, int) ([]model.Article, error)
	getFunc    func(context.Context, int64) (model.Article, error)
	updateFunc func(context.Context, int64, model.ArticleInput) (model.Article, error)
	deleteFunc func(context.Context, int64) error
}

func (stub serviceStub) Create(ctx context.Context, input model.ArticleInput) (model.Article, error) {
	return stub.createFunc(ctx, input)
}

func (stub serviceStub) List(ctx context.Context, limit, offset int) ([]model.Article, error) {
	return stub.listFunc(ctx, limit, offset)
}

func (stub serviceStub) GetByID(ctx context.Context, id int64) (model.Article, error) {
	return stub.getFunc(ctx, id)
}

func (stub serviceStub) Update(ctx context.Context, id int64, input model.ArticleInput) (model.Article, error) {
	return stub.updateFunc(ctx, id, input)
}

func (stub serviceStub) Delete(ctx context.Context, id int64) error {
	return stub.deleteFunc(ctx, id)
}

func TestArticleHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := serviceStub{
		createFunc: func(_ context.Context, input model.ArticleInput) (model.Article, error) {
			return model.Article{ID: 7, Title: input.Title, Content: input.Content, Category: input.Category, Status: input.Status}, nil
		},
	}
	router := testRouter(stub)
	body := []byte(`{"title":"A sufficiently long title","content":"content","category":"backend","status":"draft"}`)
	request := httptest.NewRequest(http.MethodPost, "/article/", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body.String())
	}
	var article model.Article
	if err := json.Unmarshal(response.Body.Bytes(), &article); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if article.ID != 7 {
		t.Errorf("response ID = %d, want 7", article.ID)
	}
}

func TestArticleHandlerValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := serviceStub{
		createFunc: func(context.Context, model.ArticleInput) (model.Article, error) {
			return model.Article{}, validator.Errors{"title": "title is required"}
		},
	}
	router := testRouter(stub)
	request := httptest.NewRequest(http.MethodPost, "/article/", bytes.NewReader([]byte(`{}`)))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestArticleHandlerNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := serviceStub{
		getFunc: func(context.Context, int64) (model.Article, error) {
			return model.Article{}, service.ErrArticleNotFound
		},
	}
	router := testRouter(stub)
	request := httptest.NewRequest(http.MethodGet, "/article/99", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func testRouter(articleService service.ArticleService) *gin.Engine {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := gin.New()
	NewArticleHandler(articleService, logger).RegisterRoutes(router)
	return router
}
