package repository

import (
	"time"

	"Web_lab/internal/users/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(token *models.Token) error {
	return r.db.Create(token).Error
}

func (r *TokenRepository) FindByHash(tokenHash string) (*models.Token, error) {
	var token models.Token
	err := r.db.Where("token_hash = ? AND revoked = false AND expires_at > ?",
		tokenHash, time.Now()).First(&token).Error
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepository) FindActiveByUser(userID uuid.UUID, tokenType string) ([]models.Token, error) {
	var tokens []models.Token
	err := r.db.
		Where("user_id = ? AND type = ? AND revoked = false AND expires_at > ?", userID, tokenType, time.Now()).
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *TokenRepository) RevokeByID(tokenID uuid.UUID) error {
	return r.db.Model(&models.Token{}).
		Where("id = ?", tokenID).
		Update("revoked", true).Error
}

func (r *TokenRepository) RevokeAll(userID uuid.UUID) error {
	return r.db.Model(&models.Token{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true).Error
}
