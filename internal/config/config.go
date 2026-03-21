package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	Port       string
}

func Load() *Config {
	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		Port:       os.Getenv("PORT"),
	}

	// Значения по умолчанию
	if cfg.DBHost == "" {
		cfg.DBHost = "localhost"
	}
	if cfg.DBPort == "" {
		cfg.DBPort = "5432"
	}
	if cfg.DBUser == "" {
		cfg.DBUser = "nix"
	}
	if cfg.DBPassword == "" {
		cfg.DBPassword = "sews"
	}
	if cfg.DBName == "" {
		cfg.DBName = "woodenButWithSoul"
	}
	if cfg.Port == "" {
		cfg.Port = "4200"
	}

	return cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}
