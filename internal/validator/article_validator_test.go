package validator

import (
	"strings"
	"testing"

	"sharing-vision-backend-test/internal/model"
)

func TestArticleValidatorValidate(t *testing.T) {
	t.Parallel()

	validInput := model.ArticleInput{
		Title:    strings.Repeat("t", minimumTitleLength),
		Content:  strings.Repeat("c", minimumContentLength),
		Category: strings.Repeat("g", minimumCategoryLength),
		Status:   "publish",
	}

	tests := []struct {
		name          string
		modify        func(*model.ArticleInput)
		expectedField string
	}{
		{
			name: "valid input",
		},
		{
			name: "missing title",
			modify: func(input *model.ArticleInput) {
				input.Title = "  "
			},
			expectedField: "title",
		},
		{
			name: "short title",
			modify: func(input *model.ArticleInput) {
				input.Title = strings.Repeat("t", minimumTitleLength-1)
			},
			expectedField: "title",
		},
		{
			name: "short content",
			modify: func(input *model.ArticleInput) {
				input.Content = strings.Repeat("c", minimumContentLength-1)
			},
			expectedField: "content",
		},
		{
			name: "short category",
			modify: func(input *model.ArticleInput) {
				input.Category = strings.Repeat("g", minimumCategoryLength-1)
			},
			expectedField: "category",
		},
		{
			name: "invalid status",
			modify: func(input *model.ArticleInput) {
				input.Status = "deleted"
			},
			expectedField: "status",
		},
	}

	articleValidator := NewArticleValidator()
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			input := validInput
			if test.modify != nil {
				test.modify(&input)
			}

			err := articleValidator.Validate(input)
			if test.expectedField == "" {
				if err != nil {
					t.Fatalf("Validate() returned unexpected error: %v", err)
				}
				return
			}

			validationErrors, ok := err.(Errors)
			if !ok {
				t.Fatalf("Validate() error type = %T, want validator.Errors", err)
			}
			if _, exists := validationErrors[test.expectedField]; !exists {
				t.Errorf("Validate() errors = %v, want field %q", validationErrors, test.expectedField)
			}
		})
	}
}

func TestArticleValidatorReturnsAllInvalidFields(t *testing.T) {
	t.Parallel()

	err := NewArticleValidator().Validate(model.ArticleInput{})
	validationErrors, ok := err.(Errors)
	if !ok {
		t.Fatalf("Validate() error type = %T, want validator.Errors", err)
	}
	if len(validationErrors) != 4 {
		t.Errorf("Validate() returned %d field errors, want 4", len(validationErrors))
	}
}
