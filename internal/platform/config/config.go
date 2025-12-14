package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr    string
	Env     string
	Version string
	DB      DBConfig
}

type DBConfig struct {
	DSN string
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
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnvOrFatal(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Fatalf("Environment variable %s is not set", key)
	return ""
}
