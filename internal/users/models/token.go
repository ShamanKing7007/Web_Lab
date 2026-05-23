package models

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID        uuid.UUID  `bson:"_id" json:"id"`
	UserID    uuid.UUID  `bson:"user_id" json:"-"`
	Type      string     `bson:"type" json:"-"`
	TokenHash string     `bson:"token_hash" json:"-"`
	ExpiresAt time.Time  `bson:"expires_at" json:"-"`
	Revoked   bool       `bson:"revoked" json:"-"`
	CreatedAt time.Time  `bson:"created_at" json:"-"`
	UpdatedAt time.Time  `bson:"updated_at" json:"-"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"-"`
}
