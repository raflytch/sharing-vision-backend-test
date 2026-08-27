package repository

import (
	"context"
	"errors"

	"sharing-vision-backend-test/internal/model"
)

// ErrArticleNotFound indicates that an article does not exist.
var ErrArticleNotFound = errors.New("article not found")

// ArticleRepository defines article persistence operations.
type ArticleRepository interface {
	Create(ctx context.Context, input model.ArticleInput) (model.Article, error)
	FindAll(ctx context.Context, limit, offset int) ([]model.Article, error)
	FindByID(ctx context.Context, id int64) (model.Article, error)
	Update(ctx context.Context, id int64, input model.ArticleInput) (model.Article, error)
	Delete(ctx context.Context, id int64) error
}
