package validator

import (
	"errors"
	"strings"
)

var (
	ErrTitleRequired = errors.New("title cannot be empty")
	ErrTitleTooLong  = errors.New("title must be at most 200 characters")
)

// Проверка заголовка заметки (1-200 символов)
func IsValidNoteTitle(title string) bool {
	title = strings.TrimSpace(title)
	return len(title) >= 1 && len(title) <= 200
}

// Проверка пагинации
func ValidatePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
