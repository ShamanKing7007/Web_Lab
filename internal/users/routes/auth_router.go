package routes

import (
	"Web_lab/internal/users/handler"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes добавляет auth роуты к существующему gin.Engine
func SetupAuthRoutes(
	router *gin.Engine,
	authHandler *handler.AuthHandler,
) {
	authMiddleware := authHandler.CreateAuthMiddleware()

	// Публичные роуты
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)

		// OAuth
		auth.GET("/oauth/:provider", authHandler.OAuthInit)
		auth.GET("/oauth/:provider/callback", authHandler.OAuthCallback)
	}

	// Защищённые роуты
	protected := router.Group("/auth")
	protected.Use(authMiddleware)
	{
		protected.GET("/whoami", authHandler.Whoami)
		protected.POST("/logout", authHandler.Logout)
		protected.POST("/logout-all", authHandler.LogoutAll)
	}
}
