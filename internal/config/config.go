package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ServerPort       string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSslMode        string
	JWTAccessSecret  string
	JWTRefreshSecret string
}

func New(log *logrus.Logger) *Config {
	if err := godotenv.Load(); err != nil {
		log.Warning("Error loading .env file")
	}

	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "admin"),
		DBName:           getEnv("DB_NAME", "music"),
		DBSslMode:        getEnv("DB_SSLMODE", "disable"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "access_secret"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "refresh_secret"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
