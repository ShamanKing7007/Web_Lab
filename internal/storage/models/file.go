package models

import (
	"time"

	"github.com/google/uuid"
)

// File stores metadata for an object kept in MinIO.
// @Description Метаданные файла, сохраненного в объектном хранилище.
type File struct {
	ID           uuid.UUID  `bson:"_id" json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID       uuid.UUID  `bson:"user_id" json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OriginalName string     `bson:"original_name" json:"original_name" example:"avatar.png"`
	ObjectKey    string     `bson:"object_key" json:"-" swaggerignore:"true"`
	Size         int64      `bson:"size" json:"size" example:"184532"`
	MimeType     string     `bson:"mime_type" json:"mime_type" example:"image/png"`
	Bucket       string     `bson:"bucket" json:"-" swaggerignore:"true"`
	CreatedAt    time.Time  `bson:"created_at" json:"created_at" example:"2026-05-23T12:00:00Z"`
	UpdatedAt    time.Time  `bson:"updated_at" json:"updated_at" example:"2026-05-23T12:00:00Z"`
	DeletedAt    *time.Time `bson:"deleted_at,omitempty" json:"-" swaggerignore:"true"`
}

// FileResponse is safe file metadata for API responses.
// @Description Публичные метаданные файла без bucket и objectKey.
type FileResponse struct {
	ID           uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OriginalName string    `json:"original_name" example:"avatar.png"`
	Size         int64     `json:"size" example:"184532"`
	MimeType     string    `json:"mime_type" example:"image/png"`
	DownloadURL  string    `json:"download_url" example:"/files/550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt    string    `json:"created_at" example:"2026-05-23T12:00:00Z"`
	UpdatedAt    string    `json:"updated_at" example:"2026-05-23T12:00:00Z"`
}

func (f *File) ToResponse() FileResponse {
	return FileResponse{
		ID:           f.ID,
		OriginalName: f.OriginalName,
		Size:         f.Size,
		MimeType:     f.MimeType,
		DownloadURL:  "/files/" + f.ID.String(),
		CreatedAt:    f.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    f.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
