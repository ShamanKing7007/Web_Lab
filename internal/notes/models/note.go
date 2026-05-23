package models

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID        uuid.UUID  `bson:"_id" json:"id"`
	UserID    *uuid.UUID `bson:"user_id,omitempty" json:"user_id"`
	Title     string     `bson:"title" json:"title"`
	Content   string     `bson:"content" json:"content"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"-"`
}
