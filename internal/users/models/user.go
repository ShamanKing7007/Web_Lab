package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User модель пользователя в системе
// @Description Пользователь с поддержкой email и OAuth (Yandex, VK)
type User struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Уникальный идентификатор пользователя"`
	Email               string         `gorm:"size:255;uniqueIndex;not null" json:"email" example:"user@example.com" description:"Email пользователя"`
	PasswordHash        string         `gorm:"size:255;not null" json:"-" description:"Хеш пароля (скрыт)"`
	Salt                string         `gorm:"size:255;not null" json:"-" description:"Соль для пароля (скрыт)"`
	YandexID            *string        `gorm:"size:100;uniqueIndex;default:null" json:"-,omitempty" description:"ID пользователя в Яндекс OAuth"`
	VKID                *string        `gorm:"size:100;uniqueIndex;default:null" json:"-,omitempty" description:"ID пользователя в VK OAuth"`
	ResetTokenHash      *string        `gorm:"size:255;default:null" json:"-" description:"Хеш токена сброса пароля (скрыт)"`
	ResetTokenExpiresAt *time.Time     `gorm:"default:null" json:"-" description:"Время истечения токена сброса пароля"`
	CreatedAt           time.Time      `json:"created_at" example:"2024-01-01T12:00:00Z" description:"Дата создания"`
	UpdatedAt           time.Time      `json:"updated_at" example:"2024-01-01T12:00:00Z" description:"Дата последнего обновления"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-" swaggerignore:"true" description:"Время мягкого удаления"`
}

// UserResponse публичные данные пользователя (без чувствительной информации)
// @Description Пользовательская информация для API ответов
type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID пользователя"`
	Email     string    `json:"email" example:"user@example.com" description:"Email пользователя"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T12:00:00Z" description:"Дата регистрации"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T12:00:00Z" description:"Дата последнего обновления"`
}

// UserWithOAuthResponse пользователь с информацией об OAuth провайдерах
// @Description Пользователь с указанием связанных OAuth аккаунтов
type UserWithOAuthResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"user@example.com"`
	HasYandex bool      `json:"has_yandex" example:"true" description:"Привязан ли Яндекс аккаунт"`
	HasVK     bool      `json:"has_vk" example:"false" description:"Привязан ли VK аккаунт"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// TableName возвращает имя таблицы в БД
func (User) TableName() string {
	return "users"
}

// IsOAuthUser проверяет, был ли пользователь создан через OAuth
func (u *User) IsOAuthUser() bool {
	return u.YandexID != nil || u.VKID != nil
}

// GetOAuthProvider возвращает название OAuth провайдера
func (u *User) GetOAuthProvider() string {
	if u.YandexID != nil {
		return "yandex"
	}
	if u.VKID != nil {
		return "vk"
	}
	return ""
}

// ToResponse преобразует модель пользователя в безопасный API-ответ.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
