package routes

import (
	"Web_lab/internal/storage/handler"

	"github.com/gin-gonic/gin"
)

func SetupFileRoutes(router *gin.Engine, fileHandler *handler.FileHandler, authMiddleware gin.HandlerFunc) {
	files := router.Group("/files")
	files.Use(authMiddleware)
	{
		files.POST("", fileHandler.UploadFile)
		files.GET("/:id", fileHandler.DownloadFile)
		files.DELETE("/:id", fileHandler.DeleteFile)
	}
}
