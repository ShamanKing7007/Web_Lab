package repository

import (
	"Web_lab/internal/database"
	"Web_lab/internal/notes/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Интерфейс репозитория для тестирования
type NoteRepository interface {
	Create(note *models.Note) error
	FindByID(id uuid.UUID) (*models.Note, error)
	FindAll(offset, limit int) ([]models.Note, int64, error)
	Update(note *models.Note) error
	Delete(id uuid.UUID) *gorm.DB
}

type NoteRepositoryImpl struct {
	db *gorm.DB
}

func NewNoteRepository(db *database.Database) NoteRepository {
	return &NoteRepositoryImpl{db: db.DB}
}

// Создание заметки
func (r *NoteRepositoryImpl) Create(note *models.Note) error {
	return r.db.Create(note).Error
}

// Поиск по ID (исключая удалённые)
func (r *NoteRepositoryImpl) FindByID(id uuid.UUID) (*models.Note, error) {
	var note models.Note
	err := r.db.First(&note, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// Получить все (с пагинацией, исключая удалённые)
func (r *NoteRepositoryImpl) FindAll(offset, limit int) ([]models.Note, int64, error) {
	var notes []models.Note
	var total int64

	r.db.Model(&models.Note{}).Count(&total)

	err := r.db.Order("created_at desc").Offset(offset).Limit(limit).Find(&notes).Error
	if err != nil {
		return nil, 0, err
	}

	return notes, total, nil
}

// Обновление
func (r *NoteRepositoryImpl) Update(note *models.Note) error {
	return r.db.Save(note).Error
}

// Soft delete — возвращает *gorm.DB для проверки RowsAffected
func (r *NoteRepositoryImpl) Delete(id uuid.UUID) *gorm.DB {
	return r.db.Delete(&models.Note{}, "id = ?", id)
}
