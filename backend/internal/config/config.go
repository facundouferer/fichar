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
	Office   OfficeConfig
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

type OfficeConfig struct {
	Latitude  float64
	Longitude float64
	RadiusKm  float64
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

	officeLat, _ := strconv.ParseFloat(getEnv("OFFICE_LATITUDE", "0"), 64)
	officeLng, _ := strconv.ParseFloat(getEnv("OFFICE_LONGITUDE", "0"), 64)
	officeRadius, _ := strconv.ParseFloat(getEnv("OFFICE_RADIUS_KM", "5"), 64)

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8082"),
		},
		Database: dbCfg,
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-me-in-production"),
		},
		Office: OfficeConfig{
			Latitude:  officeLat,
			Longitude: officeLng,
			RadiusKm:  officeRadius,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
