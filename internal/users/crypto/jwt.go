package crypto

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims — кастомные утверждения JWT
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken генерирует JWT с коротким сроком жизни (15 мин)
func GenerateAccessToken(userID uuid.UUID, secret string, expiration time.Duration) (string, error) {
	return generateToken(userID, secret, expiration)
}

// GenerateRefreshToken генерирует JWT с долгим сроком жизни (7 дней)
func GenerateRefreshToken(userID uuid.UUID, secret string, expiration time.Duration) (string, error) {
	return generateToken(userID, secret, expiration)
}

// ParseToken парсит и проверяет подпись токена
func ParseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

// generateToken — общая логика генерации JWT
func generateToken(userID uuid.UUID, secret string, expiration time.Duration) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
