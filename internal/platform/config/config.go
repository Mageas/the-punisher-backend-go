package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr    string
	Env     string
	Version string
	DB      DBConfig
	JWT     JWTConfig
}

type DBConfig struct {
	DSN string
}

type JWTConfig struct {
	AccessSecret      string
	AccessExpiration  time.Duration
	RefreshSecret     string
	RefreshExpiration time.Duration
	Issuer            string
	Audience          string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Addr:    GetEnv("APP_ADDRESS", ":8080"),
		Env:     GetEnv("APP_ENV", "development"),
		Version: GetEnv("APP_VERSION", "1.0.0"),
		DB: DBConfig{
			DSN: GetEnv("APP_DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			AccessSecret:      GetEnvOrFatal("JWT_ACCESS_SECRET"),
			AccessExpiration:  GetEnvDuration("JWT_ACCESS_EXPIRATION_IN_MINUTES", 15) * time.Minute,
			RefreshSecret:     GetEnvOrFatal("JWT_REFRESH_SECRET"),
			RefreshExpiration: GetEnvDuration("JWT_REFRESH_EXPIRATION_IN_DAYS", 7) * time.Hour * 24,
			Issuer:            GetEnvOrFatal("JWT_ISSUER"),
			Audience:          GetEnvOrFatal("JWT_AUDIENCE"),
		},
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnvDuration(key string, fallback int) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if parsedValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(parsedValue)
		}
	}
	return time.Duration(fallback)
}

func GetEnvOrFatal(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Fatalf("Environment variable %s is not set", key)
	return ""
}
