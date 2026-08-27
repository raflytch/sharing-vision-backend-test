package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"sharing-vision-backend-test/internal/model"
	"sharing-vision-backend-test/internal/repository"
	"sharing-vision-backend-test/internal/validator"
)

const maxArticleLimit = 100

var (
	// ErrArticleNotFound indicates that an article does not exist.
	ErrArticleNotFound = errors.New("article not found")
	// ErrInvalidArticleID indicates that an article ID is not positive.
	ErrInvalidArticleID = errors.New("article ID must be a positive integer")
	// ErrInvalidPagination indicates invalid list pagination values.
	ErrInvalidPagination = errors.New("limit must be between 1 and 100 and offset must be zero or greater")
)

// ArticleService defines article use cases consumed by the HTTP layer.
type ArticleService interface {
	Create(ctx context.Context, input model.ArticleInput) (model.Article, error)
	List(ctx context.Context, limit, offset int) ([]model.Article, error)
	GetByID(ctx context.Context, id int64) (model.Article, error)
	Update(ctx context.Context, id int64, input model.ArticleInput) (model.Article, error)
	Delete(ctx context.Context, id int64) error
}

type articleService struct {
	repository repository.ArticleRepository
	validator  validator.ArticleValidator
}

// NewArticleService creates an article service with explicit dependencies.
func NewArticleService(articleRepository repository.ArticleRepository, articleValidator validator.ArticleValidator) ArticleService {
	return &articleService{
		repository: articleRepository,
		validator:  articleValidator,
	}
}

func (service *articleService) Create(ctx context.Context, input model.ArticleInput) (model.Article, error) {
	input = normalizeInput(input)
	if err := service.validator.Validate(input); err != nil {
		return model.Article{}, err
	}

	article, err := service.repository.Create(ctx, input)
	if err != nil {
		return model.Article{}, fmt.Errorf("create article: %w", err)
	}
	return article, nil
}

func (service *articleService) List(ctx context.Context, limit, offset int) ([]model.Article, error) {
	if limit < 1 || limit > maxArticleLimit || offset < 0 {
		return nil, ErrInvalidPagination
	}

	articles, err := service.repository.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}
	return articles, nil
}

func (service *articleService) GetByID(ctx context.Context, id int64) (model.Article, error) {
	if id <= 0 {
		return model.Article{}, ErrInvalidArticleID
	}

	article, err := service.repository.FindByID(ctx, id)
	if errors.Is(err, repository.ErrArticleNotFound) {
		return model.Article{}, ErrArticleNotFound
	}
	if err != nil {
		return model.Article{}, fmt.Errorf("get article: %w", err)
	}
	return article, nil
}

func (service *articleService) Update(ctx context.Context, id int64, input model.ArticleInput) (model.Article, error) {
	if id <= 0 {
		return model.Article{}, ErrInvalidArticleID
	}

	input = normalizeInput(input)
	if err := service.validator.Validate(input); err != nil {
		return model.Article{}, err
	}

	article, err := service.repository.Update(ctx, id, input)
	if errors.Is(err, repository.ErrArticleNotFound) {
		return model.Article{}, ErrArticleNotFound
	}
	if err != nil {
		return model.Article{}, fmt.Errorf("update article: %w", err)
	}
	return article, nil
}

func (service *articleService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidArticleID
	}

	err := service.repository.Delete(ctx, id)
	if errors.Is(err, repository.ErrArticleNotFound) {
		return ErrArticleNotFound
	}
	if err != nil {
		return fmt.Errorf("delete article: %w", err)
	}
	return nil
}

func normalizeInput(input model.ArticleInput) model.ArticleInput {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	input.Category = strings.TrimSpace(input.Category)
	input.Status = strings.TrimSpace(input.Status)
	return input
}
