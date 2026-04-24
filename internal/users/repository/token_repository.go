package repository

import (
	"Web_lab/internal/users/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// Создание записи токена
func (r *TokenRepository) Create(token *models.Token) error {
	return r.db.Create(token).Error
}

// Поиск по хешу (активный, не отозванный, не просроченный)
func (r *TokenRepository) FindByHash(tokenHash string) (*models.Token, error) {
	var token models.Token
	err := r.db.Where("token_hash = ? AND revoked = false AND expires_at > ?",
		tokenHash, time.Now()).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// FindActiveByUser возвращает все активные refresh-токены пользователя.
func (r *TokenRepository) FindActiveByUser(userID uuid.UUID) ([]models.Token, error) {
	var tokens []models.Token
	err := r.db.
		Where("user_id = ? AND revoked = false AND expires_at > ?", userID, time.Now()).
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// RevokeByID отзывает один токен по ID.
func (r *TokenRepository) RevokeByID(tokenID uuid.UUID) error {
	return r.db.Model(&models.Token{}).
		Where("id = ?", tokenID).
		Update("revoked", true).Error
}

// Отзыв всех токенов пользователя
func (r *TokenRepository) RevokeAll(userID uuid.UUID) error {
	return r.db.Model(&models.Token{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true).Error
}
