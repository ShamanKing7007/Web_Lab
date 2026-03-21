package main

import (
	"log"

	"Web_lab/internal/config"
	"Web_lab/internal/controllers"
	"Web_lab/internal/handler"
	"Web_lab/internal/models"
	"Web_lab/internal/repository"
	"Web_lab/internal/service"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Подключение к БД
	database, err := repository.NewDatabase(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Автоматическая миграция через GORM
	err = database.DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Инициализация слоёв
	userRepo := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepo)
	userHandler := controllers.NewUserHandler(userService)

	// Info handler (старая логика)
	infoHandler := handler.NewInfoHandler()

	// Настройка роутера
	router := handler.NewRouter(userHandler, infoHandler)

	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
