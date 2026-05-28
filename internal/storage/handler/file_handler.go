package handler

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"Web_lab/internal/apperrors"
	"Web_lab/internal/httpapi"
	"Web_lab/internal/storage/models"
	"Web_lab/internal/storage/service"
	"Web_lab/internal/users/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	service     service.FileService
	maxFileSize int64
}

func NewFileHandler(svc service.FileService, maxFileSize int64) *FileHandler {
	return &FileHandler{service: svc, maxFileSize: maxFileSize}
}

// UploadFile загружает изображение текущего пользователя в MinIO.
// @Summary      Загрузить файл
// @Description  Загружает PNG/JPEG/JPG потоком через multipart/form-data и сохраняет метаданные в MongoDB
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Файл изображения PNG/JPEG/JPG"
// @Success      201 {object} models.FileResponse "Файл загружен"
// @Failure      400 {object} httpapi.ErrorResponse "Некорректный файл"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      500 {object} httpapi.ErrorResponse "Ошибка загрузки"
// @Router       /files [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxFileSize+(1<<20))
	part, err := filePart(c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "file is required"})
		return
	}
	defer part.Close()

	response, err := h.service.Upload(service.UploadInput{
		Reader:   part,
		Filename: part.FileName(),
	}, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "only png, jpeg and jpg files up to configured size are allowed"})
			return
		}

		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to upload file"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func filePart(r *http.Request) (*multipart.Part, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		if part.FormName() != "file" {
			_ = part.Close()
			continue
		}
		if part.FileName() == "" {
			_ = part.Close()
			return nil, errors.New("file part has empty filename")
		}

		return part, nil
	}
}

// DownloadFile скачивает файл текущего пользователя.
// @Summary      Скачать файл
// @Description  Возвращает поток файла из MinIO, если файл принадлежит текущему пользователю
// @Tags         files
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Param        id path string true "UUID файла" format(uuid)
// @Success      200 {file} file "Файл"
// @Failure      400 {object} httpapi.ErrorResponse "Некорректный ID"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} httpapi.ErrorResponse "Доступ к чужому файлу запрещен"
// @Failure      404 {object} httpapi.ErrorResponse "Файл не найден"
// @Failure      500 {object} httpapi.ErrorResponse "Ошибка скачивания"
// @Router       /files/{id} [get]
func (h *FileHandler) DownloadFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	download, err := h.service.GetDownload(id, userID)
	if err != nil {
		writeFileError(c, err, "failed to download file")
		return
	}
	defer download.Reader.Close()

	disposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": download.Metadata.OriginalName,
	})
	c.DataFromReader(http.StatusOK, download.Metadata.Size, download.Metadata.MimeType, download.Reader, map[string]string{
		"Content-Disposition": disposition,
	})
}

// DeleteFile удаляет файл текущего пользователя.
// @Summary      Удалить файл
// @Description  Физически удаляет объект из MinIO и помечает метаданные как удаленные
// @Tags         files
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "UUID файла" format(uuid)
// @Success      204 "Файл удален"
// @Failure      400 {object} httpapi.ErrorResponse "Некорректный ID"
// @Failure      401 {object} httpapi.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} httpapi.ErrorResponse "Доступ к чужому файлу запрещен"
// @Failure      404 {object} httpapi.ErrorResponse "Файл не найден"
// @Failure      500 {object} httpapi.ErrorResponse "Ошибка удаления"
// @Router       /files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(id, userID); err != nil {
		writeFileError(c, err, "failed to delete file")
		return
	}

	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "invalid id"})
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

func writeFileError(c *gin.Context, err error, fallback string) {
	if errors.Is(err, apperrors.ErrFileNotFound) {
		c.JSON(http.StatusNotFound, httpapi.ErrorResponse{Error: apperrors.ErrFileNotFound.Error()})
		return
	}
	if errors.Is(err, apperrors.ErrForbidden) {
		c.JSON(http.StatusForbidden, httpapi.ErrorResponse{Error: apperrors.ErrForbidden.Error()})
		return
	}

	c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: fallback})
}

var _ models.FileResponse
