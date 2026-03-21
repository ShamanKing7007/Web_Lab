package controllers

import (
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(userHandler *UserHandler) *Router {
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

	return &Router{engine: engine}
}

func (r *Router) Run(port string) error {
	return r.engine.Run(":" + port)
}
