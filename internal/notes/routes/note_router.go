package routes

import (
	"Web_lab/internal/notes/handler"

	"github.com/gin-gonic/gin"
)

type NoteRouter struct {
	Engine *gin.Engine
}

func NewNoteRouter(noteHandler *handler.NoteHandler, authMiddleware gin.HandlerFunc) *NoteRouter {
	engine := gin.Default()

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Группа маршрутов /notes (защищена middleware)
	notes := engine.Group("/notes")
	notes.Use(authMiddleware)
	{
		notes.POST("", noteHandler.CreateNote)
		notes.GET("", noteHandler.GetNotes)
		notes.GET("/:id", noteHandler.GetNote)
		notes.PUT("/:id", noteHandler.UpdateNote)
		notes.PATCH("/:id", noteHandler.PatchNote)
		notes.DELETE("/:id", noteHandler.DeleteNote)
	}

	return &NoteRouter{Engine: engine}
}

func (r *NoteRouter) Run(port string) error {
	return r.Engine.Run(":" + port)
}
