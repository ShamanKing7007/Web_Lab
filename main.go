package main

import (
	"context"
	"log"

	"Web_lab/internal/cache"
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

	_ "Web_lab/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Web Labs API
// @version         1.0
// @description     REST API для заметок с JWT-аутентификацией, refresh-сессиями и OAuth через Yandex.
// @description     Основной runtime использует HttpOnly cookies, а схема BearerAuth добавлена в Swagger UI для ручного тестирования защищенных методов.
// @termsOfService  http://swagger.io/terms/
// @license.url     https://opensource.org/licenses/MIT
// @host            localhost:4200
// @BasePath        /
// @schemes         http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите access token в формате "Bearer {token}". В приложении токен обычно передается через HttpOnly cookie access_token.

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := database.NewDatabase(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	err = db.DB.AutoMigrate(
		&noteModels.Note{},
		&userModels.User{},
		&userModels.Token{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	cacheService := cache.NewService(ctx, cache.Options{
		Host:       cfg.RedisHost,
		Port:       cfg.RedisPort,
		Password:   cfg.RedisPassword,
		DB:         cfg.RedisDB,
		DefaultTTL: cfg.CacheDefaultTTL,
	})
	defer cacheService.Close()

	noteRepo := repository.NewNoteRepository(db)
	noteService := service.NewNoteService(noteRepo, cacheService, cfg.CacheDefaultTTL)
	noteHandler := handler.NewNoteHandler(noteService)

	userRepo := userRepoPkg.NewUserRepository(db.DB)
	tokenRepo := userRepoPkg.NewTokenRepository(db.DB)

	authService := userService.NewAuthService(
		userRepo,
		tokenRepo,
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
		cacheService,
		cfg.CacheDefaultTTL,
	)

	oauthConfig := oauth.LoadOAuthConfig()
	authHandler := userHandler.NewAuthHandler(authService, userRepo, oauthConfig)

	authMiddleware := middleware.AuthMiddleware(authService.ValidateAccessToken)
	router := routes.NewNoteRouter(noteHandler, authMiddleware)
	engine := router.Engine

	if cfg.SwaggerEnabled {
		engine.GET("/api/docs/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.PersistAuthorization(true),
		))
	}

	userRoutes.SetupAuthRoutes(engine, authHandler)

	log.Printf("Server starting on port %s (env=%s, swagger=%t)", cfg.Port, cfg.AppEnv, cfg.SwaggerEnabled)
	if err := engine.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
