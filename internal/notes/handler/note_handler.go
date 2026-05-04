package handler

import (
	"errors"
	"net/http"
	"strconv"

	"Web_lab/internal/apperrors"
	"Web_lab/internal/httpapi"
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

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: apperrors.ErrInvalidID.Error()})
		return uuid.Nil, false
	}

	return id, true
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDStr, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, httpapi.ErrorResponse{Error: "unauthorized"})
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "invalid user id"})
		return uuid.Nil, false
	}

	return userID, true
}

// CreateNote создает новую заметку текущего пользователя.
// @Summary      Создать заметку
// @Description  Создает новую заметку, принадлежащую авторизованному пользователю
// @Tags         notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body service.CreateNoteDTO true "Данные заметки"
// @Success      201 {object} service.NoteResponse "Заметка создана"
// @Failure      400 {object} httpapi.ErrorResponse "Некорректные входные данные"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      500 {object} httpapi.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /notes [post]
func (h *NoteHandler) CreateNote(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var dto service.CreateNoteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: err.Error()})
		return
	}

	note, err := h.service.Create(dto, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "invalid data"})
			return
		}

		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// GetNotes возвращает заметки пользователя с пагинацией.
// @Summary      Список заметок
// @Description  Возвращает список заметок текущего пользователя с пагинацией
// @Tags         notes
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Номер страницы" default(1) minimum(1)
// @Param        limit query int false "Размер страницы" default(10) minimum(1) maximum(100)
// @Success      200 {object} service.NotesResponse "Список заметок"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      500 {object} httpapi.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /notes [get]
func (h *NoteHandler) GetNotes(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	response, err := h.service.GetAll(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to get notes"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNote возвращает одну заметку пользователя по ID.
// @Summary      Получить заметку
// @Description  Возвращает заметку по идентификатору, если она принадлежит текущему пользователю
// @Tags         notes
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID заметки" format(uuid)
// @Success      200 {object} service.NoteResponse "Заметка найдена"
// @Failure      400 {object} httpapi.ErrorResponse "Некорректный ID"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} httpapi.ErrorResponse "Доступ к чужой заметке запрещен"
// @Failure      404 {object} httpapi.ErrorResponse "Заметка не найдена"
// @Failure      500 {object} httpapi.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /notes/{id} [get]
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
			c.JSON(http.StatusNotFound, httpapi.ErrorResponse{Error: apperrors.ErrNotFound.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, httpapi.ErrorResponse{Error: apperrors.ErrForbidden.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to get note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// UpdateNote полностью обновляет заметку.
// @Summary      Полностью обновить заметку
// @Description  Полностью заменяет данные заметки текущего пользователя
// @Tags         notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID заметки" format(uuid)
// @Param        request body service.UpdateNoteDTO true "Новые данные заметки"
// @Success      200 {object} service.NoteResponse "Заметка обновлена"
// @Failure      400 {object} httpapi.ErrorResponse "Некорректные входные данные"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} httpapi.ErrorResponse "Доступ к чужой заметке запрещен"
// @Failure      404 {object} httpapi.ErrorResponse "Заметка не найдена"
// @Failure      500 {object} httpapi.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /notes/{id} [put]
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
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: err.Error()})
		return
	}

	note, err := h.service.Update(id, dto, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "invalid data"})
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, httpapi.ErrorResponse{Error: apperrors.ErrNotFound.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, httpapi.ErrorResponse{Error: apperrors.ErrForbidden.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// PatchNote частично обновляет заметку.
// @Summary      Частично обновить заметку
// @Description  Частично обновляет поля заметки текущего пользователя
// @Tags         notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID заметки" format(uuid)
// @Param        request body service.UpdateNoteDTO true "Изменяемые поля заметки"
// @Success      200 {object} service.NoteResponse "Заметка обновлена"
// @Failure      400 {object} httpapi.ErrorResponse "Некорректные входные данные"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} httpapi.ErrorResponse "Доступ к чужой заметке запрещен"
// @Failure      404 {object} httpapi.ErrorResponse "Заметка не найдена"
// @Failure      500 {object} httpapi.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /notes/{id} [patch]
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
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: err.Error()})
		return
	}

	note, err := h.service.Patch(id, dto, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "invalid data"})
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, httpapi.ErrorResponse{Error: apperrors.ErrNotFound.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, httpapi.ErrorResponse{Error: apperrors.ErrForbidden.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote выполняет soft delete заметки.
// @Summary      Удалить заметку
// @Description  Помечает заметку как удаленную, не удаляя запись физически из базы
// @Tags         notes
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID заметки" format(uuid)
// @Success      204 "Заметка удалена"
// @Failure      400 {object} httpapi.ErrorResponse "Некорректный ID"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} httpapi.ErrorResponse "Доступ к чужой заметке запрещен"
// @Failure      404 {object} httpapi.ErrorResponse "Заметка не найдена"
// @Failure      500 {object} httpapi.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /notes/{id} [delete]
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
			c.JSON(http.StatusNotFound, httpapi.ErrorResponse{Error: apperrors.ErrNotFound.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, httpapi.ErrorResponse{Error: apperrors.ErrForbidden.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to delete note"})
		return
	}

	c.Status(http.StatusNoContent)
}
