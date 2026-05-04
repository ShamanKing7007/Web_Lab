package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Token struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"-"`
	Type      string         `gorm:"size:20;not null;index" json:"-"`
	TokenHash string         `gorm:"size:255;not null;index" json:"-"`
	ExpiresAt time.Time      `gorm:"not null" json:"-"`
	Revoked   bool           `gorm:"default:false" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
