package models

import (
	"time"

	"github.com/google/uuid"
)

// User stores authentication data in MongoDB.
// @Description Пользователь с поддержкой email и OAuth (Yandex, VK)
type User struct {
	ID                  uuid.UUID  `bson:"_id" json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Уникальный идентификатор пользователя"`
	Email               string     `bson:"email" json:"email" example:"user@example.com" description:"Email пользователя"`
	DisplayName         *string    `bson:"display_name,omitempty" json:"display_name,omitempty" example:"Ivan Petrov" description:"Отображаемое имя"`
	Bio                 *string    `bson:"bio,omitempty" json:"bio,omitempty" example:"Backend developer" description:"Описание профиля"`
	AvatarFileID        *uuid.UUID `bson:"avatar_file_id,omitempty" json:"avatar_file_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID файла аватара"`
	PasswordHash        string     `bson:"password_hash" json:"-" description:"Хеш пароля (скрыт)"`
	Salt                string     `bson:"salt" json:"-" description:"Соль для пароля (скрыт)"`
	YandexID            *string    `bson:"yandex_id,omitempty" json:"-,omitempty" description:"ID пользователя в Яндекс OAuth"`
	VKID                *string    `bson:"vk_id,omitempty" json:"-,omitempty" description:"ID пользователя в VK OAuth"`
	ResetTokenHash      *string    `bson:"reset_token_hash,omitempty" json:"-" description:"Хеш токена сброса пароля (скрыт)"`
	ResetTokenExpiresAt *time.Time `bson:"reset_token_expires_at,omitempty" json:"-" description:"Время истечения токена сброса пароля"`
	CreatedAt           time.Time  `bson:"created_at" json:"created_at" example:"2024-01-01T12:00:00Z" description:"Дата создания"`
	UpdatedAt           time.Time  `bson:"updated_at" json:"updated_at" example:"2024-01-01T12:00:00Z" description:"Дата последнего обновления"`
	DeletedAt           *time.Time `bson:"deleted_at,omitempty" json:"-" swaggerignore:"true" description:"Время мягкого удаления"`
}

// UserResponse contains public user data for API responses.
// @Description Пользовательская информация для API ответов
type UserResponse struct {
	ID           uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID пользователя"`
	Email        string     `json:"email" example:"user@example.com" description:"Email пользователя"`
	DisplayName  *string    `json:"display_name,omitempty" example:"Ivan Petrov" description:"Отображаемое имя"`
	Bio          *string    `json:"bio,omitempty" example:"Backend developer" description:"Описание профиля"`
	AvatarFileID *uuid.UUID `json:"avatar_file_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID файла аватара"`
	CreatedAt    time.Time  `json:"created_at" example:"2024-01-01T12:00:00Z" description:"Дата регистрации"`
	UpdatedAt    time.Time  `json:"updated_at" example:"2024-01-01T12:00:00Z" description:"Дата последнего обновления"`
}

// UserWithOAuthResponse shows linked OAuth providers.
// @Description Пользователь с указанием связанных OAuth аккаунтов
type UserWithOAuthResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"user@example.com"`
	HasYandex bool      `json:"has_yandex" example:"true" description:"Привязан ли Яндекс аккаунт"`
	HasVK     bool      `json:"has_vk" example:"false" description:"Привязан ли VK аккаунт"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// IsOAuthUser checks whether the user was created through OAuth.
func (u *User) IsOAuthUser() bool {
	return u.YandexID != nil || u.VKID != nil
}

// GetOAuthProvider returns the linked OAuth provider name.
func (u *User) GetOAuthProvider() string {
	if u.YandexID != nil {
		return "yandex"
	}
	if u.VKID != nil {
		return "vk"
	}
	return ""
}

// ToResponse converts the user model to a safe API response.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		DisplayName:  u.DisplayName,
		Bio:          u.Bio,
		AvatarFileID: u.AvatarFileID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
