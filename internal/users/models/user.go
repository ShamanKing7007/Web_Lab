package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email               string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash        string         `gorm:"size:255;not null" json:"-"`
	Salt                string         `gorm:"size:255;not null" json:"-"`
	YandexID            *string        `gorm:"size:100;uniqueIndex;default:null" json:"-"`
	VKID                *string        `gorm:"size:100;uniqueIndex;default:null" json:"-"`
	ResetTokenHash      *string        `gorm:"size:255;default:null" json:"-"`
	ResetTokenExpiresAt *time.Time     `gorm:"default:null" json:"-"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}
