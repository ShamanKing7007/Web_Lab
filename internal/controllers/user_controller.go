package controllers

import (
	"errors"
	"net/http"
	"strconv"

	apperrors "Web_lab/internal/errors"
	"Web_lab/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// CreateUser создаёт нового пользователя
func (h *UserHandler) CreateUser(c *gin.Context) {
	var dto service.CreateUserDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Create(dto)
	if err != nil {
		if errors.Is(err, apperrors.ErrAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUsers получает всех пользователей с пагинацией
func (h *UserHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	response, err := h.service.GetAll(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUser получает пользователя по ID
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidID.Error()})
		return
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser полностью обновляет пользователя
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidID.Error()})
		return
	}

	var dto service.UpdateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Update(id, dto)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// PatchUser частично обновляет пользователя
func (h *UserHandler) PatchUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidID.Error()})
		return
	}

	var dto service.UpdateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Patch(id, dto)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser удаляет пользователя (Soft Delete)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidID.Error()})
		return
	}

	err = h.service.Delete(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": apperrors.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}
