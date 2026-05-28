package handler

import (
	"errors"
	"net/http"

	"Web_lab/internal/apperrors"
	"Web_lab/internal/httpapi"
	"Web_lab/internal/users/dto"
	"Web_lab/internal/users/middleware"
	"Web_lab/internal/users/models"
	"Web_lab/internal/users/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	service service.ProfileService
}

func NewProfileHandler(svc service.ProfileService) *ProfileHandler {
	return &ProfileHandler{service: svc}
}

// GetProfile возвращает профиль текущего пользователя.
// @Summary      Получить профиль
// @Description  Возвращает профиль текущего авторизованного пользователя
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.UserResponse "Профиль пользователя"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      404 {object} httpapi.ErrorResponse "Пользователь не найден"
// @Router       /profile [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, ok := getProfileUserID(c)
	if !ok {
		return
	}

	response, err := h.service.Get(userID)
	if err != nil {
		writeProfileError(c, err, "failed to get profile")
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProfile обновляет профиль текущего пользователя.
// @Summary      Обновить профиль
// @Description  Обновляет displayName, bio и avatarFileId после проверки владения файлом
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.UpdateProfileDTO true "Данные профиля"
// @Success      200 {object} models.UserResponse "Профиль обновлен"
// @Failure      400 {object} httpapi.ErrorResponse "Некорректные данные"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} httpapi.ErrorResponse "Файл принадлежит другому пользователю"
// @Failure      404 {object} httpapi.ErrorResponse "Пользователь или файл не найден"
// @Failure      500 {object} httpapi.ErrorResponse "Ошибка обновления"
// @Router       /profile [post]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, ok := getProfileUserID(c)
	if !ok {
		return
	}

	var req dto.UpdateProfileDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := h.service.Update(userID, req)
	if err != nil {
		writeProfileError(c, err, "failed to update profile")
		return
	}

	c.JSON(http.StatusOK, response)
}

func getProfileUserID(c *gin.Context) (uuid.UUID, bool) {
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

func writeProfileError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, apperrors.ErrUserNotFound):
		c.JSON(http.StatusNotFound, httpapi.ErrorResponse{Error: apperrors.ErrUserNotFound.Error()})
	case errors.Is(err, apperrors.ErrFileNotFound):
		c.JSON(http.StatusNotFound, httpapi.ErrorResponse{Error: apperrors.ErrFileNotFound.Error()})
	case errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, httpapi.ErrorResponse{Error: apperrors.ErrForbidden.Error()})
	case errors.Is(err, apperrors.ErrValidation):
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "invalid profile data"})
	default:
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: fallback})
	}
}

var _ models.UserResponse
