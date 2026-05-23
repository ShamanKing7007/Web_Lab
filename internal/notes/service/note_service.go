package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"Web_lab/internal/apperrors"
	"Web_lab/internal/cache"
	"Web_lab/internal/notes/models"
	"Web_lab/internal/notes/repository"
	"Web_lab/internal/notes/validator"

	"github.com/google/uuid"
)

// NoteService бизнес-логика заметок.
type NoteService interface {
	Create(dto CreateNoteDTO, userID uuid.UUID) (*NoteResponse, error)
	GetByID(id uuid.UUID, userID uuid.UUID) (*NoteResponse, error)
	GetAll(userID uuid.UUID, page, limit int) (*NotesResponse, error)
	Update(id uuid.UUID, dto UpdateNoteDTO, userID uuid.UUID) (*NoteResponse, error)
	Patch(id uuid.UUID, dto UpdateNoteDTO, userID uuid.UUID) (*NoteResponse, error)
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type NoteServiceImpl struct {
	repo     repository.NoteRepository
	cache    *cache.Service
	cacheTTL time.Duration
}

func NewNoteService(repo repository.NoteRepository, cacheService *cache.Service, cacheTTL time.Duration) NoteService {
	return &NoteServiceImpl{
		repo:     repo,
		cache:    cacheService,
		cacheTTL: cacheTTL,
	}
}

// CreateNoteDTO DTO для создания заметки.
type CreateNoteDTO struct {
	Title   string `json:"title" binding:"required,min=1,max=200" example:"Первая заметка"`
	Content string `json:"content" binding:"omitempty" example:"Текст заметки"`
}

// UpdateNoteDTO DTO для обновления заметки.
type UpdateNoteDTO struct {
	Title   string `json:"title" binding:"omitempty,min=1,max=200" example:"Обновленная заметка"`
	Content string `json:"content" binding:"omitempty" example:"Обновленный текст"`
}

// NoteResponse публичное представление заметки.
type NoteResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title     string    `json:"title" example:"Первая заметка"`
	Content   string    `json:"content" example:"Текст заметки"`
	CreatedAt string    `json:"created_at" example:"2026-05-02T12:00:00Z"`
	UpdatedAt string    `json:"updated_at" example:"2026-05-02T12:10:00Z"`
}

// PaginationMeta мета-информация пагинации.
type PaginationMeta struct {
	Total      int64 `json:"total" example:"10"`
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"10"`
	TotalPages int   `json:"total_pages" example:"1"`
}

// NotesResponse ответ со списком заметок.
type NotesResponse struct {
	Data []NoteResponse `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

func toNoteResponse(note *models.Note) *NoteResponse {
	return &NoteResponse{
		ID:        note.ID,
		Title:     note.Title,
		Content:   note.Content,
		CreatedAt: note.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: note.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toNoteResponses(notes []models.Note) []NoteResponse {
	responses := make([]NoteResponse, 0, len(notes))
	for i := range notes {
		responses = append(responses, *toNoteResponse(&notes[i]))
	}

	return responses
}

func (s *NoteServiceImpl) Create(dto CreateNoteDTO, userID uuid.UUID) (*NoteResponse, error) {
	if !validator.IsValidNoteTitle(dto.Title) {
		return nil, apperrors.ErrValidation
	}

	note := &models.Note{
		ID:      uuid.New(),
		UserID:  &userID,
		Title:   dto.Title,
		Content: dto.Content,
	}

	if err := s.repo.Create(note); err != nil {
		return nil, err
	}

	s.invalidateNoteListCache(userID)

	return toNoteResponse(note), nil
}

func (s *NoteServiceImpl) GetByID(id uuid.UUID, userID uuid.UUID) (*NoteResponse, error) {
	ctx := context.Background()
	key := noteDetailKey(userID, id)

	var cached NoteResponse
	if ok, err := s.cache.Get(ctx, key, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		log.Printf("failed to read note cache %s: %v", key, err)
	}

	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperrors.ErrNotFound
	}

	if note.UserID == nil || *note.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	response := toNoteResponse(note)
	if err := s.cache.Set(ctx, key, response, s.cacheTTL); err != nil {
		log.Printf("failed to write note cache %s: %v", key, err)
	}

	return response, nil
}

func (s *NoteServiceImpl) GetAll(userID uuid.UUID, page, limit int) (*NotesResponse, error) {
	page, limit = validator.ValidatePagination(page, limit)
	ctx := context.Background()
	key := noteListKey(userID, page, limit)

	var cached NotesResponse
	if ok, err := s.cache.Get(ctx, key, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		log.Printf("failed to read notes list cache %s: %v", key, err)
	}

	offset := (page - 1) * limit

	notes, total, err := s.repo.FindAllByUser(userID, offset, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	response := &NotesResponse{
		Data: toNoteResponses(notes),
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}

	if err := s.cache.Set(ctx, key, response, s.cacheTTL); err != nil {
		log.Printf("failed to write notes list cache %s: %v", key, err)
	}

	return response, nil
}

func (s *NoteServiceImpl) Update(id uuid.UUID, dto UpdateNoteDTO, userID uuid.UUID) (*NoteResponse, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperrors.ErrNotFound
	}

	if note.UserID == nil || *note.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	if dto.Title == "" || !validator.IsValidNoteTitle(dto.Title) {
		return nil, apperrors.ErrValidation
	}

	note.Title = dto.Title
	note.Content = dto.Content

	if err := s.repo.Update(note); err != nil {
		return nil, err
	}

	s.invalidateNoteCache(userID, id)
	s.invalidateNoteListCache(userID)

	return toNoteResponse(note), nil
}

func (s *NoteServiceImpl) Patch(id uuid.UUID, dto UpdateNoteDTO, userID uuid.UUID) (*NoteResponse, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperrors.ErrNotFound
	}

	if note.UserID == nil || *note.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	if dto.Title != "" && !validator.IsValidNoteTitle(dto.Title) {
		return nil, apperrors.ErrValidation
	}

	if dto.Title != "" {
		note.Title = dto.Title
	}
	if dto.Content != "" {
		note.Content = dto.Content
	}

	if err := s.repo.Update(note); err != nil {
		return nil, err
	}

	s.invalidateNoteCache(userID, id)
	s.invalidateNoteListCache(userID)

	return toNoteResponse(note), nil
}

func (s *NoteServiceImpl) Delete(id uuid.UUID, userID uuid.UUID) error {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return apperrors.ErrNotFound
	}

	if note.UserID == nil || *note.UserID != userID {
		return apperrors.ErrForbidden
	}

	deleted, err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	if !deleted {
		return apperrors.ErrNotFound
	}

	s.invalidateNoteCache(userID, id)
	s.invalidateNoteListCache(userID)

	return nil
}

func (s *NoteServiceImpl) invalidateNoteCache(userID, noteID uuid.UUID) {
	key := noteDetailKey(userID, noteID)
	if err := s.cache.Del(context.Background(), key); err != nil {
		log.Printf("failed to invalidate note cache %s: %v", key, err)
	}
}

func (s *NoteServiceImpl) invalidateNoteListCache(userID uuid.UUID) {
	pattern := noteListPattern(userID)
	if err := s.cache.DelByPattern(context.Background(), pattern); err != nil {
		log.Printf("failed to invalidate notes list cache %s: %v", pattern, err)
	}
}

func noteListKey(userID uuid.UUID, page, limit int) string {
	return fmt.Sprintf("wp:notes:user:%s:list:page:%d:limit:%d", userID, page, limit)
}

func noteListPattern(userID uuid.UUID) string {
	return fmt.Sprintf("wp:notes:user:%s:list:*", userID)
}

func noteDetailKey(userID, noteID uuid.UUID) string {
	return fmt.Sprintf("wp:notes:user:%s:detail:%s", userID, noteID)
}
