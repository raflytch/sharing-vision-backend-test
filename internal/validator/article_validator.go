package validator

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"sharing-vision-backend-test/internal/model"
)

const (
	minimumTitleLength    = 20
	minimumContentLength  = 200
	minimumCategoryLength = 3
)

var validStatuses = map[string]struct{}{
	"publish": {},
	"draft":   {},
	"thrash":  {},
}

// Errors contains validation messages keyed by request field.
type Errors map[string]string

func (Errors) Error() string {
	return "article validation failed"
}

// ArticleValidator defines article input validation behavior.
type ArticleValidator interface {
	Validate(input model.ArticleInput) error
}

type articleValidator struct{}

// NewArticleValidator creates an article validator.
func NewArticleValidator() ArticleValidator {
	return articleValidator{}
}

func (articleValidator) Validate(input model.ArticleInput) error {
	errors := make(Errors)

	validateRequiredLength(errors, "title", input.Title, minimumTitleLength)
	validateRequiredLength(errors, "content", input.Content, minimumContentLength)
	validateRequiredLength(errors, "category", input.Category, minimumCategoryLength)

	status := strings.TrimSpace(input.Status)
	if status == "" {
		errors["status"] = "status is required"
	} else if _, valid := validStatuses[status]; !valid {
		errors["status"] = "status must be one of: publish, draft, thrash"
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

func validateRequiredLength(errors Errors, field, value string, minimum int) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		errors[field] = field + " is required"
		return
	}
	if utf8.RuneCountInString(trimmed) < minimum {
		errors[field] = field + " must be at least " + strconv.Itoa(minimum) + " characters"
	}
}
