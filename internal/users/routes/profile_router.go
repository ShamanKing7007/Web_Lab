package routes

import (
	"Web_lab/internal/users/handler"

	"github.com/gin-gonic/gin"
)

func SetupProfileRoutes(router *gin.Engine, profileHandler *handler.ProfileHandler, authMiddleware gin.HandlerFunc) {
	profile := router.Group("/profile")
	profile.Use(authMiddleware)
	{
		profile.GET("", profileHandler.GetProfile)
		profile.POST("", profileHandler.UpdateProfile)
	}
}
