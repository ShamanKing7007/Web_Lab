package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv           string
	SwaggerEnabled   bool
	DBName           string
	MongoURI         string
	Port             string
	RedisHost        string
	RedisPort        string
	RedisPassword    string
	RedisDB          int
	CacheDefaultTTL  time.Duration
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
	MinIOEndpoint    string
	MinIOAccessKey   string
	MinIOSecretKey   string
	MinIOBucket      string
	MinIOUseSSL      bool
	MaxFileSize      int64
	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQUser     string
	RabbitMQPass     string
	RabbitMQExchange string
	RabbitMQDLX      string
	UserRegisteredQ  string
	SMTPHost         string
	SMTPPort         string
	SMTPUser         string
	SMTPPass         string
	SMTPFrom         string
	SMTPSecure       bool
}

func Load() *Config {
	appEnv := resolveAppEnv()

	cfg := &Config{
		AppEnv:           appEnv,
		SwaggerEnabled:   resolveSwaggerEnabled(appEnv),
		DBName:           getEnvOrDefault("DB_NAME", "Web_Labs"),
		MongoURI:         os.Getenv("MONGO_URI"),
		Port:             getEnvOrDefault("PORT", "4200"),
		RedisHost:        getEnvOrDefault("REDIS_HOST", "localhost"),
		RedisPort:        getEnvOrDefault("REDIS_PORT", "6379"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		RedisDB:          mustParseIntEnv("REDIS_DB", "0"),
		CacheDefaultTTL:  mustParseCacheTTL("CACHE_TTL_DEFAULT", "300"),
		JWTAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
		JWTAccessTTL:     mustParseDurationEnv("JWT_ACCESS_EXPIRATION", "15m"),
		JWTRefreshTTL:    mustParseDurationEnv("JWT_REFRESH_EXPIRATION", "7d"),
		MinIOEndpoint:    os.Getenv("MINIO_ENDPOINT"),
		MinIOAccessKey:   os.Getenv("MINIO_ACCESS_KEY"),
		MinIOSecretKey:   os.Getenv("MINIO_SECRET_KEY"),
		MinIOBucket:      getEnvOrDefault("MINIO_BUCKET", "web-labs-files"),
		MinIOUseSSL:      mustParseBoolEnv("MINIO_USE_SSL", "false"),
		MaxFileSize:      mustParseInt64Env("MAX_FILE_SIZE", "10485760"),
		RabbitMQHost:     getEnvOrDefault("RABBITMQ_HOST", "localhost"),
		RabbitMQPort:     getEnvOrDefault("RABBITMQ_PORT", "5672"),
		RabbitMQUser:     os.Getenv("RABBITMQ_USER"),
		RabbitMQPass:     os.Getenv("RABBITMQ_PASS"),
		RabbitMQExchange: getEnvOrDefault("RABBITMQ_EXCHANGE", "app.events"),
		RabbitMQDLX:      getEnvOrDefault("RABBITMQ_DLX", "app.dlx"),
		UserRegisteredQ:  getEnvOrDefault("QUEUE_USER_REGISTERED", "wp.auth.user.registered"),
		SMTPHost:         os.Getenv("SMTP_HOST"),
		SMTPPort:         getEnvOrDefault("SMTP_PORT", "465"),
		SMTPUser:         os.Getenv("SMTP_USER"),
		SMTPPass:         os.Getenv("SMTP_PASS"),
		SMTPFrom:         os.Getenv("SMTP_FROM"),
		SMTPSecure:       mustParseBoolEnv("SMTP_SECURE", "true"),
	}

	if cfg.MongoURI == "" {
		log.Fatal("MONGO_URI environment variable is required")
	}
	if cfg.JWTAccessSecret == "" {
		log.Fatal("JWT_ACCESS_SECRET environment variable is required")
	}
	if cfg.JWTRefreshSecret == "" {
		log.Fatal("JWT_REFRESH_SECRET environment variable is required")
	}
	if cfg.RedisPassword == "" {
		log.Fatal("REDIS_PASSWORD environment variable is required")
	}
	if cfg.MinIOEndpoint == "" {
		log.Fatal("MINIO_ENDPOINT environment variable is required")
	}
	if cfg.MinIOAccessKey == "" {
		log.Fatal("MINIO_ACCESS_KEY environment variable is required")
	}
	if cfg.MinIOSecretKey == "" {
		log.Fatal("MINIO_SECRET_KEY environment variable is required")
	}
	if cfg.RabbitMQUser == "" {
		log.Fatal("RABBITMQ_USER environment variable is required")
	}
	if cfg.RabbitMQPass == "" {
		log.Fatal("RABBITMQ_PASS environment variable is required")
	}
	if cfg.SMTPHost == "" {
		log.Fatal("SMTP_HOST environment variable is required")
	}
	if cfg.SMTPUser == "" {
		log.Fatal("SMTP_USER environment variable is required")
	}
	if cfg.SMTPPass == "" {
		log.Fatal("SMTP_PASS environment variable is required")
	}
	if cfg.SMTPFrom == "" {
		log.Fatal("SMTP_FROM environment variable is required")
	}

	return cfg
}

func mustParseIntEnv(key, defaultVal string) int {
	raw := getEnvOrDefault(key, defaultVal)
	value, err := strconv.Atoi(raw)
	if err != nil {
		log.Fatalf("%s has invalid integer value %q", key, raw)
	}

	return value
}

func mustParseInt64Env(key, defaultVal string) int64 {
	raw := getEnvOrDefault(key, defaultVal)
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		log.Fatalf("%s has invalid integer value %q", key, raw)
	}

	return value
}

func mustParseBoolEnv(key, defaultVal string) bool {
	raw := getEnvOrDefault(key, defaultVal)
	value, err := strconv.ParseBool(raw)
	if err != nil {
		log.Fatalf("%s has invalid boolean value %q", key, raw)
	}

	return value
}

func mustParseCacheTTL(key, defaultVal string) time.Duration {
	raw := getEnvOrDefault(key, defaultVal)
	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second
	}

	duration, err := parseDurationWithDays(raw)
	if err != nil {
		log.Fatalf("%s has invalid duration %q: %v", key, raw, err)
	}

	return duration
}

func mustParseDurationEnv(key, defaultVal string) time.Duration {
	raw := getEnvOrDefault(key, defaultVal)
	duration, err := parseDurationWithDays(raw)
	if err != nil {
		log.Fatalf("%s has invalid duration %q: %v", key, raw, err)
	}

	return duration
}

func parseDurationWithDays(value string) (time.Duration, error) {
	if strings.HasSuffix(value, "d") {
		daysPart := strings.TrimSuffix(value, "d")
		days, err := strconv.Atoi(daysPart)
		if err != nil {
			return 0, fmt.Errorf("invalid day count")
		}

		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(value)
}

func resolveAppEnv() string {
	if value := os.Getenv("APP_ENV"); value != "" {
		return strings.ToLower(value)
	}
	if value := os.Getenv("NODE_ENV"); value != "" {
		return strings.ToLower(value)
	}

	return "development"
}

func resolveSwaggerEnabled(appEnv string) bool {
	if raw := os.Getenv("SWAGGER_ENABLED"); raw != "" {
		enabled, err := strconv.ParseBool(raw)
		if err != nil {
			log.Fatalf("SWAGGER_ENABLED has invalid boolean value %q", raw)
		}

		return enabled
	}

	return appEnv != "production"
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return defaultVal
}
