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
	userHandler "Web_lab/internal/users/handler"
	"Web_lab/internal/users/middleware"
	userModels "Web_lab/internal/users/models"
	"Web_lab/internal/users/oauth"
	userRepoPkg "Web_lab/internal/users/repository"
	userRoutes "Web_lab/internal/users/routes"
	userService "Web_lab/internal/users/service"
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
	err = db.DB.AutoMigrate(
		&noteModels.Note{},
		&userModels.User{},
		&userModels.Token{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// === NOTES слои ===
	noteRepo := repository.NewNoteRepository(db)
	noteService := service.NewNoteService(noteRepo)
	noteHandler := handler.NewNoteHandler(noteService)

	// === USERS слои ===
	userRepo := userRepoPkg.NewUserRepository(db.DB)
	tokenRepo := userRepoPkg.NewTokenRepository(db.DB)

	authService := userService.NewAuthService(
		userRepo,
		tokenRepo,
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
	)

	oauthConfig := oauth.LoadOAuthConfig()
	authHandler := userHandler.NewAuthHandler(authService, userRepo, oauthConfig)

	// === Роутер ===
	authMiddleware := middleware.AuthMiddleware(cfg.JWTAccessSecret)
	router := routes.NewNoteRouter(noteHandler, authMiddleware)
	engine := router.Engine

	// Auth роуты
	userRoutes.SetupAuthRoutes(engine, authHandler, cfg.JWTAccessSecret)

	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.Port)
	if err := engine.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
