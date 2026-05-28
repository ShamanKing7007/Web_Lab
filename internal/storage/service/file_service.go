package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"Web_lab/internal/apperrors"
	"Web_lab/internal/cache"
	"Web_lab/internal/storage/models"
	"Web_lab/internal/storage/repository"

	"github.com/google/uuid"
)

const sniffSize = 512

type FileService interface {
	Upload(input UploadInput, userID uuid.UUID) (*models.FileResponse, error)
	GetDownload(id uuid.UUID, userID uuid.UUID) (*DownloadFile, error)
	GetMetadata(id uuid.UUID, userID uuid.UUID) (*models.File, error)
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type FileServiceImpl struct {
	repo        repository.FileRepository
	objectStore ObjectStorage
	cache       *cache.Service
	bucket      string
	maxFileSize int64
	cacheTTL    time.Duration
}

type DownloadFile struct {
	Metadata *models.File
	Reader   io.ReadCloser
}

type UploadInput struct {
	Reader   io.Reader
	Filename string
}

type cachedFileMetadata struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	OriginalName string     `json:"original_name"`
	ObjectKey    string     `json:"object_key"`
	Size         int64      `json:"size"`
	MimeType     string     `json:"mime_type"`
	Bucket       string     `json:"bucket"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

func NewFileService(
	repo repository.FileRepository,
	objectStore ObjectStorage,
	cacheService *cache.Service,
	bucket string,
	maxFileSize int64,
	cacheTTL time.Duration,
) FileService {
	return &FileServiceImpl{
		repo:        repo,
		objectStore: objectStore,
		cache:       cacheService,
		bucket:      bucket,
		maxFileSize: maxFileSize,
		cacheTTL:    cacheTTL,
	}
}

func (s *FileServiceImpl) Upload(input UploadInput, userID uuid.UUID) (*models.FileResponse, error) {
	if input.Reader == nil {
		return nil, apperrors.ErrValidation
	}

	mimeType, reader, err := sniffReader(input.Reader)
	if err != nil {
		return nil, err
	}
	if !isAllowedImageMime(mimeType) {
		return nil, apperrors.ErrValidation
	}

	id := uuid.New()
	objectKey := objectKeyFor(userID, id, mimeType)
	limitedReader := &maxSizeReader{
		reader: reader,
		max:    s.maxFileSize,
	}
	metadata := &models.File{
		ID:           id,
		UserID:       userID,
		OriginalName: CleanFilename(input.Filename),
		ObjectKey:    objectKey,
		MimeType:     mimeType,
		Bucket:       s.bucket,
	}

	ctx := context.Background()
	if err := s.objectStore.Upload(ctx, objectKey, mimeType, -1, limitedReader); err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			return nil, apperrors.ErrValidation
		}
		return nil, err
	}
	metadata.Size = limitedReader.BytesRead()

	if err := s.repo.Create(metadata); err != nil {
		if deleteErr := s.objectStore.Delete(ctx, objectKey); deleteErr != nil {
			log.Printf("failed to cleanup uploaded object %s after metadata error: %v", objectKey, deleteErr)
		}
		return nil, err
	}

	response := metadata.ToResponse()
	if err := s.cache.Set(ctx, fileMetaKey(id), fileToCache(metadata), s.cacheTTL); err != nil {
		log.Printf("failed to write file metadata cache %s: %v", fileMetaKey(id), err)
	}

	return &response, nil
}

func (s *FileServiceImpl) GetDownload(id uuid.UUID, userID uuid.UUID) (*DownloadFile, error) {
	metadata, err := s.GetMetadata(id, userID)
	if err != nil {
		return nil, err
	}

	exists, err := s.objectStore.Exists(context.Background(), metadata.ObjectKey)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.ErrFileNotFound
	}

	reader, err := s.objectStore.Download(context.Background(), metadata.ObjectKey)
	if err != nil {
		return nil, err
	}

	return &DownloadFile{Metadata: metadata, Reader: reader}, nil
}

func (s *FileServiceImpl) GetMetadata(id uuid.UUID, userID uuid.UUID) (*models.File, error) {
	ctx := context.Background()
	key := fileMetaKey(id)

	var cached cachedFileMetadata
	if ok, err := s.cache.Get(ctx, key, &cached); err == nil && ok {
		if cached.UserID != userID {
			return nil, apperrors.ErrForbidden
		}
		return cached.toFile(), nil
	} else if err != nil {
		log.Printf("failed to read file metadata cache %s: %v", key, err)
	}

	file, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperrors.ErrFileNotFound
	}
	if file.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	if err := s.cache.Set(ctx, key, fileToCache(file), s.cacheTTL); err != nil {
		log.Printf("failed to write file metadata cache %s: %v", key, err)
	}

	return file, nil
}

func (s *FileServiceImpl) Delete(id uuid.UUID, userID uuid.UUID) error {
	file, err := s.GetMetadata(id, userID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	deleted, err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	if !deleted {
		return apperrors.ErrFileNotFound
	}

	s.invalidateFileCache(id)
	if err := s.objectStore.Delete(ctx, file.ObjectKey); err != nil {
		log.Printf("failed to delete object %s from storage after metadata soft delete: %v", file.ObjectKey, err)
	}

	return nil
}

func sniffReader(reader io.Reader) (string, io.Reader, error) {
	buffer := make([]byte, sniffSize)
	n, err := reader.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", nil, err
	}
	if n == 0 {
		return "", nil, apperrors.ErrValidation
	}

	mimeType := http.DetectContentType(buffer[:n])
	stream := io.MultiReader(bytes.NewReader(buffer[:n]), reader)
	return mimeType, stream, nil
}

type maxSizeReader struct {
	reader io.Reader
	max    int64
	read   int64
}

func (r *maxSizeReader) Read(p []byte) (int, error) {
	if r.read == r.max {
		var probe [1]byte
		n, err := r.reader.Read(probe[:])
		if n > 0 {
			return 0, apperrors.ErrValidation
		}
		return 0, err
	}

	remaining := r.max - r.read
	if remaining < int64(len(p)) {
		p = p[:remaining]
	}

	n, err := r.reader.Read(p)
	r.read += int64(n)
	return n, err
}

func (r *maxSizeReader) BytesRead() int64 {
	return r.read
}

func isAllowedImageMime(mimeType string) bool {
	return mimeType == "image/png" || mimeType == "image/jpeg" || mimeType == "image/jpg"
}

func objectKeyFor(userID, fileID uuid.UUID, mimeType string) string {
	ext := ".jpg"
	if mimeType == "image/png" {
		ext = ".png"
	}

	return fmt.Sprintf("users/%s/files/%s%s", userID, fileID, ext)
}

func (s *FileServiceImpl) invalidateFileCache(id uuid.UUID) {
	key := fileMetaKey(id)
	if err := s.cache.Del(context.Background(), key); err != nil {
		log.Printf("failed to invalidate file cache %s: %v", key, err)
	}
}

func fileMetaKey(id uuid.UUID) string {
	return fmt.Sprintf("wp:files:%s:meta", id)
}

func fileToCache(file *models.File) cachedFileMetadata {
	return cachedFileMetadata{
		ID:           file.ID,
		UserID:       file.UserID,
		OriginalName: file.OriginalName,
		ObjectKey:    file.ObjectKey,
		Size:         file.Size,
		MimeType:     file.MimeType,
		Bucket:       file.Bucket,
		CreatedAt:    file.CreatedAt,
		UpdatedAt:    file.UpdatedAt,
		DeletedAt:    file.DeletedAt,
	}
}

func (m cachedFileMetadata) toFile() *models.File {
	return &models.File{
		ID:           m.ID,
		UserID:       m.UserID,
		OriginalName: m.OriginalName,
		ObjectKey:    m.ObjectKey,
		Size:         m.Size,
		MimeType:     m.MimeType,
		Bucket:       m.Bucket,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		DeletedAt:    m.DeletedAt,
	}
}

func CleanFilename(name string) string {
	cleaned := filepath.Base(strings.TrimSpace(name))
	if cleaned == "." || cleaned == string(filepath.Separator) || cleaned == "" {
		return "file"
	}

	return cleaned
}
