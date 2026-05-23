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
