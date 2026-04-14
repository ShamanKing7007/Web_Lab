package main

import (
	"log"

	"Web_lab/internal/config"
	"Web_lab/internal/database"
	"Web_lab/internal/notes/handler"
	noteModels "Web_lab/internal/notes/models"
	"Web_lab/internal/notes/repository"
	"Web_lab/internal/notes/routes"
	"Web_lab/internal/notes/service"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Подключение к БД
	db, err := database.NewDatabase(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Автоматическая миграция через GORM
	err = db.DB.AutoMigrate(&noteModels.Note{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Инициализация слоёв (Dependency Injection)
	noteRepo := repository.NewNoteRepository(db)
	noteService := service.NewNoteService(noteRepo)
	noteHandler := handler.NewNoteHandler(noteService)

	// Настройка роутера
	router := routes.NewNoteRouter(noteHandler)

	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
