package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"sharing-vision-backend-test/internal/model"
	"sharing-vision-backend-test/internal/repository"
	"sharing-vision-backend-test/internal/validator"
)

type repositoryStub struct {
	createFunc   func(context.Context, model.ArticleInput) (model.Article, error)
	findAllFunc  func(context.Context, int, int) ([]model.Article, error)
	findByIDFunc func(context.Context, int64) (model.Article, error)
	updateFunc   func(context.Context, int64, model.ArticleInput) (model.Article, error)
	deleteFunc   func(context.Context, int64) error
}

func (stub repositoryStub) Create(ctx context.Context, input model.ArticleInput) (model.Article, error) {
	return stub.createFunc(ctx, input)
}

func (stub repositoryStub) FindAll(ctx context.Context, limit, offset int) ([]model.Article, error) {
	return stub.findAllFunc(ctx, limit, offset)
}

func (stub repositoryStub) FindByID(ctx context.Context, id int64) (model.Article, error) {
	return stub.findByIDFunc(ctx, id)
}

func (stub repositoryStub) Update(ctx context.Context, id int64, input model.ArticleInput) (model.Article, error) {
	return stub.updateFunc(ctx, id, input)
}

func (stub repositoryStub) Delete(ctx context.Context, id int64) error {
	return stub.deleteFunc(ctx, id)
}

func TestArticleServiceCreateNormalizesInput(t *testing.T) {
	t.Parallel()

	input := validArticleInput()
	input.Title = "  " + input.Title + "  "

	repository := repositoryStub{
		createFunc: func(_ context.Context, received model.ArticleInput) (model.Article, error) {
			if received.Title != strings.TrimSpace(input.Title) {
				t.Errorf("Create() title = %q, want trimmed title", received.Title)
			}
			return model.Article{ID: 1, Title: received.Title}, nil
		},
	}

	articleService := NewArticleService(repository, validator.NewArticleValidator())
	article, err := articleService.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}
	if article.ID != 1 {
		t.Errorf("Create() ID = %d, want 1", article.ID)
	}
}

func TestArticleServiceCreateRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	repositoryCalled := false
	repository := repositoryStub{
		createFunc: func(context.Context, model.ArticleInput) (model.Article, error) {
			repositoryCalled = true
			return model.Article{}, nil
		},
	}

	articleService := NewArticleService(repository, validator.NewArticleValidator())
	_, err := articleService.Create(context.Background(), model.ArticleInput{})
	if err == nil {
		t.Fatal("Create() error = nil, want validation error")
	}
	if repositoryCalled {
		t.Error("Create() called repository for invalid input")
	}
}

func TestArticleServiceListValidatesPagination(t *testing.T) {
	t.Parallel()

	articleService := NewArticleService(repositoryStub{}, validator.NewArticleValidator())
	tests := []struct {
		limit  int
		offset int
	}{
		{limit: 0, offset: 0},
		{limit: maxArticleLimit + 1, offset: 0},
		{limit: 10, offset: -1},
	}

	for _, test := range tests {
		_, err := articleService.List(context.Background(), test.limit, test.offset)
		if !errors.Is(err, ErrInvalidPagination) {
			t.Errorf("List(%d, %d) error = %v, want ErrInvalidPagination", test.limit, test.offset, err)
		}
	}
}

func TestArticleServiceTranslatesNotFoundErrors(t *testing.T) {
	t.Parallel()

	repository := repositoryStub{
		findByIDFunc: func(context.Context, int64) (model.Article, error) {
			return model.Article{}, repository.ErrArticleNotFound
		},
		updateFunc: func(context.Context, int64, model.ArticleInput) (model.Article, error) {
			return model.Article{}, repository.ErrArticleNotFound
		},
		deleteFunc: func(context.Context, int64) error {
			return repository.ErrArticleNotFound
		},
	}
	articleService := NewArticleService(repository, validator.NewArticleValidator())

	if _, err := articleService.GetByID(context.Background(), 1); !errors.Is(err, ErrArticleNotFound) {
		t.Errorf("GetByID() error = %v, want ErrArticleNotFound", err)
	}
	if _, err := articleService.Update(context.Background(), 1, validArticleInput()); !errors.Is(err, ErrArticleNotFound) {
		t.Errorf("Update() error = %v, want ErrArticleNotFound", err)
	}
	if err := articleService.Delete(context.Background(), 1); !errors.Is(err, ErrArticleNotFound) {
		t.Errorf("Delete() error = %v, want ErrArticleNotFound", err)
	}
}

func validArticleInput() model.ArticleInput {
	return model.ArticleInput{
		Title:    strings.Repeat("t", 20),
		Content:  strings.Repeat("c", 200),
		Category: "backend",
		Status:   "draft",
	}
}
