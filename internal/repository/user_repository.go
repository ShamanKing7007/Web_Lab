package repository

import (
	"Web_lab/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(database *Database) *UserRepository {
	return &UserRepository{db: database.DB}
}

// Создание пользователя
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Поиск по ID (исключая удалённые)
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Получить всех (с пагинацией, исключая удалённых)
func (r *UserRepository) FindAll(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Получаем общее количество
	r.db.Model(&models.User{}).Count(&total)

	// Получаем данные с пагинацией
	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Обновление
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Soft delete
func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}
