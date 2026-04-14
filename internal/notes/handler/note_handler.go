package handler

import (
	"errors"
	"net/http"
	"strconv"

	"Web_lab/internal/apperrors"
	"Web_lab/internal/notes/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NoteHandler struct {
	service service.NoteService
}

func NewNoteHandler(svc service.NoteService) *NoteHandler {
	return &NoteHandler{service: svc}
}

// parseID — хелпер для парсинга UUID из параметра запроса
func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidID.Error()})
		return uuid.Nil, false
	}
	return id, true
}

// CreateNote создаёт новую заметку
func (h *NoteHandler) CreateNote(c *gin.Context) {
	var dto service.CreateNoteDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.Create(dto)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// GetNotes получает все заметки с пагинацией
func (h *NoteHandler) GetNotes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	response, err := h.service.GetAll(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notes"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNote получает заметку по ID
func (h *NoteHandler) GetNote(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	note, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// UpdateNote полностью обновляет заметку (PUT)
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var dto service.UpdateNoteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.Update(id, dto)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// PatchNote частично обновляет заметку (PATCH)
func (h *NoteHandler) PatchNote(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var dto service.UpdateNoteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.Patch(id, dto)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote удаляет заметку (Soft Delete)
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	err := h.service.Delete(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete note"})
		return
	}

	c.Status(http.StatusNoContent)
}
