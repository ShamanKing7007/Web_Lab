package repository

import (
	"Web_lab/internal/users/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Создание пользователя
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Поиск по email (исключая удалённые)
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Поиск по ID (исключая удалённые)
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Поиск по Yandex ID (исключая удалённые)
func (r *UserRepository) FindByYandexID(yandexID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("yandex_id = ?", yandexID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUsersWithActiveResetToken возвращает пользователей с неистёкшими reset-токенами.
func (r *UserRepository) FindUsersWithActiveResetToken() ([]models.User, error) {
	var users []models.User
	err := r.db.
		Where("reset_token_hash IS NOT NULL AND reset_token_expires_at IS NOT NULL AND reset_token_expires_at > ?", time.Now()).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Обновление пользователя
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}
