package service

import (
	"Web_lab/internal/apperrors"
	"Web_lab/internal/notes/models"
	"Web_lab/internal/notes/repository"
	"Web_lab/internal/notes/validator"

	"github.com/google/uuid"
)

// Интерфейс сервиса для тестирования
type NoteService interface {
	Create(dto CreateNoteDTO, userID uuid.UUID) (*models.Note, error)
	GetByID(id uuid.UUID, userID uuid.UUID) (*models.Note, error)
	GetAll(userID uuid.UUID, page, limit int) (*NotesResponse, error)
	Update(id uuid.UUID, dto UpdateNoteDTO, userID uuid.UUID) (*models.Note, error)
	Patch(id uuid.UUID, dto UpdateNoteDTO, userID uuid.UUID) (*models.Note, error)
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type NoteServiceImpl struct {
	repo repository.NoteRepository
}

func NewNoteService(repo repository.NoteRepository) NoteService {
	return &NoteServiceImpl{repo: repo}
}

// DTO для создания заметки
type CreateNoteDTO struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"omitempty"`
}

// DTO для обновления заметки
type UpdateNoteDTO struct {
	Title   string `json:"title" binding:"omitempty,min=1,max=200"`
	Content string `json:"content" binding:"omitempty"`
}

// Мета для пагинации
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// Ответ со списком заметок
type NotesResponse struct {
	Data []models.Note  `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// Создание заметки (привязка к пользователю)
func (s *NoteServiceImpl) Create(dto CreateNoteDTO, userID uuid.UUID) (*models.Note, error) {
	if !validator.IsValidNoteTitle(dto.Title) {
		return nil, apperrors.ErrValidation
	}

	note := &models.Note{
		ID:      uuid.New(),
		UserID:  &userID,
		Title:   dto.Title,
		Content: dto.Content,
	}

	err := s.repo.Create(note)
	if err != nil {
		return nil, err
	}

	return note, nil
}

// Поиск по ID (с проверкой владения)
func (s *NoteServiceImpl) GetByID(id uuid.UUID, userID uuid.UUID) (*models.Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperrors.ErrNotFound
	}

	if note.UserID == nil || *note.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	return note, nil
}

// Получить все заметки пользователя с пагинацией
func (s *NoteServiceImpl) GetAll(userID uuid.UUID, page, limit int) (*NotesResponse, error) {
	page, limit = validator.ValidatePagination(page, limit)
	offset := (page - 1) * limit

	notes, total, err := s.repo.FindAllByUser(userID, offset, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &NotesResponse{
		Data: notes,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// Полное обновление заметки (PUT) — title обязателен + проверка владения
func (s *NoteServiceImpl) Update(id uuid.UUID, dto UpdateNoteDTO, userID uuid.UUID) (*models.Note, error) {
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

	err = s.repo.Update(note)
	if err != nil {
		return nil, err
	}

	return note, nil
}

// Частичное обновление заметки (PATCH) + проверка владения
func (s *NoteServiceImpl) Patch(id uuid.UUID, dto UpdateNoteDTO, userID uuid.UUID) (*models.Note, error) {
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

	err = s.repo.Update(note)
	if err != nil {
		return nil, err
	}

	return note, nil
}

// Soft delete — с проверкой владения
func (s *NoteServiceImpl) Delete(id uuid.UUID, userID uuid.UUID) error {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return apperrors.ErrNotFound
	}

	if note.UserID == nil || *note.UserID != userID {
		return apperrors.ErrForbidden
	}

	result := s.repo.Delete(id)
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	if result.Error != nil {
		return result.Error
	}
	return nil
}
