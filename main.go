package main

import (
	"context"
	"log"

	"Web_lab/internal/cache"
	"Web_lab/internal/config"
	"Web_lab/internal/database"
	"Web_lab/internal/notes/handler"
	"Web_lab/internal/notes/repository"
	"Web_lab/internal/notes/routes"
	"Web_lab/internal/notes/service"
	storageHandler "Web_lab/internal/storage/handler"
	storageRepo "Web_lab/internal/storage/repository"
	storageRoutes "Web_lab/internal/storage/routes"
	storageService "Web_lab/internal/storage/service"
	userHandler "Web_lab/internal/users/handler"
	"Web_lab/internal/users/middleware"
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
// @contact.name   API Support
// @contact.email  support@weblab.com
// @license.name   MIT
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

	db, err := database.NewDatabase(ctx, cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(context.Background())
	log.Println("MongoDB connection and indexes initialized successfully")

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

	userRepo := userRepoPkg.NewUserRepository(db)
	tokenRepo := userRepoPkg.NewTokenRepository(db)
	fileRepo := storageRepo.NewFileRepository(db)

	objectStorage, err := storageService.NewMinIOStorage(ctx, storageService.MinIOOptions{
		Endpoint:  cfg.MinIOEndpoint,
		AccessKey: cfg.MinIOAccessKey,
		SecretKey: cfg.MinIOSecretKey,
		Bucket:    cfg.MinIOBucket,
		UseSSL:    cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO storage: %v", err)
	}
	fileService := storageService.NewFileService(
		fileRepo,
		objectStorage,
		cacheService,
		cfg.MinIOBucket,
		cfg.MaxFileSize,
		cfg.CacheDefaultTTL,
	)
	fileHandler := storageHandler.NewFileHandler(fileService, cfg.MaxFileSize)

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
	profileService := userService.NewProfileService(userRepo, fileService, cacheService, cfg.CacheDefaultTTL)
	profileHandler := userHandler.NewProfileHandler(profileService)

	authMiddleware := middleware.AuthMiddleware(authService.ValidateAccessToken)
	router := routes.NewNoteRouter(noteHandler, authMiddleware)
	engine := router.Engine
	engine.MaxMultipartMemory = cfg.MaxFileSize

	if cfg.SwaggerEnabled {
		engine.GET("/api/docs/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.PersistAuthorization(true),
		))
	}

	userRoutes.SetupAuthRoutes(engine, authHandler)
	userRoutes.SetupProfileRoutes(engine, profileHandler, authMiddleware)
	storageRoutes.SetupFileRoutes(engine, fileHandler, authMiddleware)

	log.Printf("Server starting on port %s (env=%s, swagger=%t)", cfg.Port, cfg.AppEnv, cfg.SwaggerEnabled)
	if err := engine.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
