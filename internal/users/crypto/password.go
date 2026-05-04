package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword генерирует уникальную соль и хеширует пароль
func HashPassword(password string) (hash string, salt string, err error) {
	salt = generateSalt()
	normalized := normalizeSecret(salt + password)

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(normalized), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	return string(hashedBytes), salt, nil
}

// CheckPassword проверяет соответствие пароля хешу с учётом соли
func CheckPassword(password, hash, salt string) bool {
	normalized := normalizeSecret(salt + password)
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(normalized)) == nil {
		return true
	}

	// Совместимость со старыми записями, созданными до нормализации длинных секретов.
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(salt+password)) == nil
}

// HashOpaqueToken хеширует токен без внешней соли.
func HashOpaqueToken(token string) (string, error) {
	normalized := normalizeOpaqueToken(token)

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(normalized), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// CheckOpaqueToken сравнивает токен с bcrypt-хешем.
func CheckOpaqueToken(token, hash string) bool {
	normalized := normalizeOpaqueToken(token)
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(normalized)) == nil
}

// GenerateOpaqueToken генерирует случайный токен для одноразовых сценариев.
func GenerateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// generateSalt генерирует уникальную соль
func generateSalt() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func normalizeOpaqueToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeSecret(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
