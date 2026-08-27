package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"sharing-vision-backend-test/internal/model"
	"sharing-vision-backend-test/internal/service"
	"sharing-vision-backend-test/internal/validator"
)

type ArticleHandler struct {
	service service.ArticleService
	logger  *slog.Logger
}

// NewArticleHandler creates an HTTP handler for article endpoints.
func NewArticleHandler(articleService service.ArticleService, logger *slog.Logger) *ArticleHandler {
	return &ArticleHandler{service: articleService, logger: logger}
}

// RegisterRoutes registers all article API routes.
func (handler *ArticleHandler) RegisterRoutes(router *gin.Engine) {
	articles := router.Group("/article")
	articles.POST("/", handler.Create)
	articles.GET("/:idOrLimit/:offset", handler.List)
	articles.GET("/:idOrLimit", handler.GetByID)
	articles.PUT("/:idOrLimit", handler.Update)
	articles.PATCH("/:idOrLimit", handler.Update)
	articles.DELETE("/:idOrLimit", handler.Delete)
}

func (handler *ArticleHandler) Create(context *gin.Context) {
	input, ok := bindArticleInput(context)
	if !ok {
		return
	}

	article, err := handler.service.Create(context.Request.Context(), input)
	if err != nil {
		handler.writeError(context, err)
		return
	}
	context.JSON(http.StatusCreated, article)
}

func (handler *ArticleHandler) List(context *gin.Context) {
	limit, err := strconv.Atoi(context.Param("idOrLimit"))
	if err != nil {
		writeJSONError(context, http.StatusBadRequest, "limit must be an integer")
		return
	}
	offset, err := strconv.Atoi(context.Param("offset"))
	if err != nil {
		writeJSONError(context, http.StatusBadRequest, "offset must be an integer")
		return
	}

	articles, err := handler.service.List(context.Request.Context(), limit, offset)
	if err != nil {
		handler.writeError(context, err)
		return
	}
	context.JSON(http.StatusOK, articles)
}

func (handler *ArticleHandler) GetByID(context *gin.Context) {
	id, ok := parseArticleID(context)
	if !ok {
		return
	}

	article, err := handler.service.GetByID(context.Request.Context(), id)
	if err != nil {
		handler.writeError(context, err)
		return
	}
	context.JSON(http.StatusOK, article)
}

func (handler *ArticleHandler) Update(context *gin.Context) {
	id, ok := parseArticleID(context)
	if !ok {
		return
	}
	input, ok := bindArticleInput(context)
	if !ok {
		return
	}

	article, err := handler.service.Update(context.Request.Context(), id, input)
	if err != nil {
		handler.writeError(context, err)
		return
	}
	context.JSON(http.StatusOK, article)
}

func (handler *ArticleHandler) Delete(context *gin.Context) {
	id, ok := parseArticleID(context)
	if !ok {
		return
	}

	if err := handler.service.Delete(context.Request.Context(), id); err != nil {
		handler.writeError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "article deleted successfully"})
}

func bindArticleInput(context *gin.Context) (model.ArticleInput, bool) {
	var input model.ArticleInput
	if err := context.ShouldBindJSON(&input); err != nil {
		writeJSONError(context, http.StatusBadRequest, "request body must contain valid JSON")
		return model.ArticleInput{}, false
	}
	return input, true
}

func parseArticleID(context *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(context.Param("idOrLimit"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(context, http.StatusBadRequest, "article ID must be a positive integer")
		return 0, false
	}
	return id, true
}

func (handler *ArticleHandler) writeError(context *gin.Context, err error) {
	var validationErrors validator.Errors
	switch {
	case errors.As(err, &validationErrors):
		context.JSON(http.StatusBadRequest, gin.H{
			"error":  "validation failed",
			"fields": validationErrors,
		})
	case errors.Is(err, service.ErrInvalidArticleID), errors.Is(err, service.ErrInvalidPagination):
		writeJSONError(context, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrArticleNotFound):
		writeJSONError(context, http.StatusNotFound, err.Error())
	default:
		handler.logger.Error(
			"article request failed",
			"method", context.Request.Method,
			"path", context.Request.URL.Path,
			"error", err,
		)
		writeJSONError(context, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSONError(context *gin.Context, status int, message string) {
	context.JSON(status, gin.H{"error": message})
}
