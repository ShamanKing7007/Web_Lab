package validator

import (
	"net/mail"
	"regexp"
	"strings"
)

// Проверка email
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Проверка пароля (минимум 4 символа)
func IsValidPassword(password string) bool {
	return len(password) >= 4
}

// Проверка имени (2-100 символов, не пустое)
func IsValidName(name string) bool {
	name = strings.TrimSpace(name)
	return len(name) >= 2 && len(name) <= 100
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

// Проверка UUID
func IsValidUUID(id string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(id))
}
