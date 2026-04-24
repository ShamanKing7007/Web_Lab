package handler

import (
	"errors"
	"net/http"
	"strconv"

	"Web_lab/internal/apperrors"
	"Web_lab/internal/notes/service"
	"Web_lab/internal/users/middleware"

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

// getUserID — извлекает userID из контекста (устанавливается middleware)
func getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDStr, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return uuid.Nil, false
	}

	return userID, true
}

// CreateNote создаёт новую заметку (привязка к пользователю)
func (h *NoteHandler) CreateNote(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var dto service.CreateNoteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.Create(dto, userID)
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

// GetNotes получает все заметки пользователя с пагинацией
func (h *NoteHandler) GetNotes(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	response, err := h.service.GetAll(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notes"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNote получает заметку по ID (с проверкой владения)
func (h *NoteHandler) GetNote(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	note, err := h.service.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": apperrors.ErrForbidden.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// UpdateNote полностью обновляет заметку (PUT) — только владелец
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	var dto service.UpdateNoteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.Update(id, dto, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": apperrors.ErrForbidden.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// PatchNote частично обновляет заметку (PATCH) — только владелец
func (h *NoteHandler) PatchNote(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	var dto service.UpdateNoteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.Patch(id, dto, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": apperrors.ErrForbidden.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote удаляет заметку (Soft Delete) — только владелец
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	err := h.service.Delete(id, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": apperrors.ErrForbidden.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete note"})
		return
	}

	c.Status(http.StatusNoContent)
}
