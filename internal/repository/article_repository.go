package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sharing-vision-backend-test/internal/model"
)

const (
	createArticleQuery = `
		INSERT INTO posts (title, content, category, status)
		VALUES (?, ?, ?, ?)`
	listArticlesQuery = `
		SELECT id, title, content, category, status
		FROM posts
		ORDER BY created_date DESC, id DESC
		LIMIT ? OFFSET ?`
	findArticleByIDQuery = `
		SELECT id, title, content, category, status
		FROM posts
		WHERE id = ?
		LIMIT 1`
	updateArticleQuery = `
		UPDATE posts
		SET title = ?, content = ?, category = ?, status = ?, updated_date = NOW()
		WHERE id = ?`
	deleteArticleQuery = `DELETE FROM posts WHERE id = ?`
)

type mysqlArticleRepository struct {
	db           *sql.DB
	queryTimeout time.Duration
}

// NewMySQLArticleRepository creates a raw SQL article repository.
func NewMySQLArticleRepository(db *sql.DB, queryTimeout time.Duration) ArticleRepository {
	return &mysqlArticleRepository{db: db, queryTimeout: queryTimeout}
}

func (repository *mysqlArticleRepository) Create(ctx context.Context, input model.ArticleInput) (model.Article, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()

	result, err := repository.db.ExecContext(
		queryContext,
		createArticleQuery,
		input.Title,
		input.Content,
		input.Category,
		input.Status,
	)
	if err != nil {
		return model.Article{}, fmt.Errorf("insert article: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Article{}, fmt.Errorf("read inserted article ID: %w", err)
	}

	return articleFromInput(id, input), nil
}

func (repository *mysqlArticleRepository) FindAll(ctx context.Context, limit, offset int) ([]model.Article, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()

	rows, err := repository.db.QueryContext(queryContext, listArticlesQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query articles: %w", err)
	}
	defer rows.Close()

	articles := make([]model.Article, 0, limit)
	for rows.Next() {
		var article model.Article
		if err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.Category, &article.Status); err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		articles = append(articles, article)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate articles: %w", err)
	}

	return articles, nil
}

func (repository *mysqlArticleRepository) FindByID(ctx context.Context, id int64) (model.Article, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()

	var article model.Article
	err := repository.db.QueryRowContext(queryContext, findArticleByIDQuery, id).Scan(
		&article.ID,
		&article.Title,
		&article.Content,
		&article.Category,
		&article.Status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Article{}, ErrArticleNotFound
	}
	if err != nil {
		return model.Article{}, fmt.Errorf("query article by ID: %w", err)
	}

	return article, nil
}

func (repository *mysqlArticleRepository) Update(ctx context.Context, id int64, input model.ArticleInput) (model.Article, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()

	result, err := repository.db.ExecContext(
		queryContext,
		updateArticleQuery,
		input.Title,
		input.Content,
		input.Category,
		input.Status,
		id,
	)
	if err != nil {
		return model.Article{}, fmt.Errorf("update article: %w", err)
	}

	matchedRows, err := result.RowsAffected()
	if err != nil {
		return model.Article{}, fmt.Errorf("read updated row count: %w", err)
	}
	if matchedRows == 0 {
		return model.Article{}, ErrArticleNotFound
	}

	return articleFromInput(id, input), nil
}

func (repository *mysqlArticleRepository) Delete(ctx context.Context, id int64) error {
	queryContext, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()

	result, err := repository.db.ExecContext(queryContext, deleteArticleQuery, id)
	if err != nil {
		return fmt.Errorf("delete article: %w", err)
	}

	deletedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted row count: %w", err)
	}
	if deletedRows == 0 {
		return ErrArticleNotFound
	}

	return nil
}

func articleFromInput(id int64, input model.ArticleInput) model.Article {
	return model.Article{
		ID:       id,
		Title:    input.Title,
		Content:  input.Content,
		Category: input.Category,
		Status:   input.Status,
	}
}
