package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword генерирует уникальную соль и хеширует пароль
func HashPassword(password string) (hash string, salt string, err error) {
	salt = generateSalt()

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(salt+password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	return string(hashedBytes), salt, nil
}

// CheckPassword проверяет соответствие пароля хешу с учётом соли
func CheckPassword(password, hash, salt string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(salt+password))
	return err == nil
}

// HashOpaqueToken хеширует токен без внешней соли.
func HashOpaqueToken(token string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// CheckOpaqueToken сравнивает токен с bcrypt-хешем.
func CheckOpaqueToken(token, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(token)) == nil
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
