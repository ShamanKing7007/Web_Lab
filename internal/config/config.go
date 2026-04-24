package config

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	Port             string
	JWTAccessSecret  string
	JWTRefreshSecret string
}

func Load() *Config {
	cfg := &Config{
		DBHost:           getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:           getEnvOrDefault("DB_PORT", "5432"),
		DBUser:           getEnvOrDefault("DB_USER", ""),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBName:           getEnvOrDefault("DB_NAME", "Web_Labs"),
		Port:             getEnvOrDefault("PORT", "4200"),
		JWTAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
	}

	// Обязательные параметры
	if cfg.DBUser == "" {
		log.Fatal("DB_USER environment variable is required")
	}
	if cfg.DBPassword == "" {
		log.Fatal("DB_PASSWORD environment variable is required")
	}
	if cfg.JWTAccessSecret == "" {
		log.Fatal("JWT_ACCESS_SECRET environment variable is required")
	}
	if cfg.JWTRefreshSecret == "" {
		log.Fatal("JWT_REFRESH_SECRET environment variable is required")
	}

	return cfg
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// DSN — формат для GORM
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}
