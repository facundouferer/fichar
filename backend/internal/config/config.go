package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	Secret string
}

func Load() (*Config, error) {
	dbURL := getEnv("DATABASE_URL", "")

	// Parse DATABASE_URL if provided, otherwise use individual vars
	dbCfg := DatabaseConfig{
		URL: dbURL,
	}

	if dbURL == "" {
		dbCfg.Host = getEnv("DB_HOST", "localhost")
		dbCfg.Port, _ = strconv.Atoi(getEnv("DB_PORT", "5432"))
		dbCfg.User = getEnv("DB_USER", "fichar_user")
		dbCfg.Password = getEnv("DB_PASSWORD", "changeme")
		dbCfg.DBName = getEnv("DB_NAME", "fichar")
		dbCfg.URL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.DBName)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: dbCfg,
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-me-in-production"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
