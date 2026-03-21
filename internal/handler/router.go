package handler

import (
	"Web_lab/internal/controllers"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(userHandler *controllers.UserHandler, infoHandler *InfoHandler) *Router {
	engine := gin.Default()

	// Группа маршрутов /users
	users := engine.Group("/users")
	{
		users.POST("", userHandler.CreateUser)
		users.GET("", userHandler.GetUsers)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.PATCH("/:id", userHandler.PatchUser)
		users.DELETE("/:id", userHandler.DeleteUser)
	}

	// Маршрут /info
	engine.GET("/info", infoHandler.InfoHandler)

	return &Router{engine: engine}
}

func (r *Router) Run(port string) error {
	return r.engine.Run(":" + port)
}
