package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr          string
	Env           string
	Version       string
	AllowRegister bool
	CORS          CORSConfig
	DB            DBConfig
	JWT           JWTConfig
	SMTP          SMTPConfig
	EmailConfirm  EmailConfirmationConfig
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type DBConfig struct {
	DSN string
}

type JWTConfig struct {
	AccessSecret        string
	AccessExpiration    time.Duration
	RefreshSecret       string
	RefreshExpiration   time.Duration
	RefreshCookieSecure bool
	Issuer              string
	Audience            string
}

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

type EmailConfirmationConfig struct {
	Secret     string
	Expiration time.Duration
	BaseURL    string
	Issuer     string
	Audience   string
}

func Load() *Config {
	godotenv.Load()

	appEnv := GetEnv("APP_ENV", "development")
	jwtIssuer := GetEnvOrFatal("JWT_ISSUER")
	jwtAudience := GetEnvOrFatal("JWT_AUDIENCE")

	cfg := &Config{
		Addr:          GetEnv("APP_ADDRESS", ":8080"),
		Env:           appEnv,
		Version:       GetEnv("APP_VERSION", "1.0.0"),
		AllowRegister: GetEnvBool("APP_ALLOW_REGISTER", true),
		CORS: CORSConfig{
			AllowedOrigins:   GetEnvCSV("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
			AllowedMethods:   GetEnvCSV("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   GetEnvCSV("CORS_ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}),
			ExposedHeaders:   GetEnvCSV("CORS_EXPOSED_HEADERS", []string{"Link"}),
			AllowCredentials: GetEnvBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           GetEnvInt("CORS_MAX_AGE", 300),
		},
		DB: DBConfig{
			DSN: GetEnv("APP_DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			AccessSecret:        GetEnvOrFatal("JWT_ACCESS_SECRET"),
			AccessExpiration:    GetEnvDuration("JWT_ACCESS_EXPIRATION_IN_MINUTES", 15) * time.Minute,
			RefreshSecret:       GetEnvOrFatal("JWT_REFRESH_SECRET"),
			RefreshExpiration:   GetEnvDuration("JWT_REFRESH_EXPIRATION_IN_DAYS", 7) * time.Hour * 24,
			RefreshCookieSecure: GetEnvBool("JWT_REFRESH_COOKIE_SECURE", strings.EqualFold(appEnv, "production")),
			Issuer:              jwtIssuer,
			Audience:            jwtAudience,
		},
		SMTP: SMTPConfig{
			Host:      GetEnv("SMTP_HOST", "localhost"),
			Port:      GetEnvInt("SMTP_PORT", 1025),
			Username:  GetEnv("SMTP_USERNAME", ""),
			Password:  GetEnv("SMTP_PASSWORD", ""),
			FromEmail: GetEnv("SMTP_FROM_EMAIL", "no-reply@the-punisher.local"),
			FromName:  GetEnv("SMTP_FROM_NAME", "The Punisher"),
		},
		EmailConfirm: EmailConfirmationConfig{
			Secret:     GetEnv("EMAIL_CONFIRMATION_SECRET", GetEnvOrFatal("JWT_ACCESS_SECRET")),
			Expiration: GetEnvDuration("EMAIL_CONFIRMATION_EXPIRATION_IN_HOURS", 24) * time.Hour,
			BaseURL:    GetEnv("EMAIL_CONFIRMATION_BASE_URL", "http://localhost:8080/v1/auth/confirm-email"),
			Issuer:     jwtIssuer,
			Audience:   jwtAudience,
		},
	}

	validateCORSConfig(cfg.CORS)
	validateSMTPConfig(cfg.SMTP)
	validateEmailConfirmationConfig(cfg.EmailConfirm)
	return cfg
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if parsedValue, err := strconv.ParseBool(value); err == nil {
			return parsedValue
		}
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

func GetEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if parsedValue, err := strconv.Atoi(value); err == nil {
			return parsedValue
		}
	}
	return fallback
}

func GetEnvCSV(key string, fallback []string) []string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return fallback
	}

	return values
}

func GetEnvOrFatal(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Fatalf("Environment variable %s is not set", key)
	return ""
}

func validateCORSConfig(cors CORSConfig) {
	if len(cors.AllowedOrigins) == 0 {
		log.Fatal("CORS_ALLOWED_ORIGINS must contain at least one origin")
	}

	if cors.MaxAge < 0 {
		log.Fatal("CORS_MAX_AGE must be >= 0")
	}

	if cors.AllowCredentials {
		for _, origin := range cors.AllowedOrigins {
			if strings.Contains(origin, "*") {
				log.Fatalf("invalid CORS config: wildcard origin %q cannot be used when CORS_ALLOW_CREDENTIALS=true", origin)
			}
		}
	}
}

func validateSMTPConfig(smtp SMTPConfig) {
	if strings.TrimSpace(smtp.Host) == "" {
		log.Fatal("SMTP_HOST must not be empty")
	}

	if smtp.Port <= 0 {
		log.Fatal("SMTP_PORT must be > 0")
	}

	if strings.TrimSpace(smtp.FromEmail) == "" {
		log.Fatal("SMTP_FROM_EMAIL must not be empty")
	}
}

func validateEmailConfirmationConfig(emailConfig EmailConfirmationConfig) {
	if strings.TrimSpace(emailConfig.Secret) == "" {
		log.Fatal("EMAIL_CONFIRMATION_SECRET must not be empty")
	}

	if emailConfig.Expiration <= 0 {
		log.Fatal("EMAIL_CONFIRMATION_EXPIRATION_IN_HOURS must be > 0")
	}

	baseURL := strings.TrimSpace(emailConfig.BaseURL)
	if baseURL == "" {
		log.Fatal("EMAIL_CONFIRMATION_BASE_URL must not be empty")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil || !parsedURL.IsAbs() || strings.TrimSpace(parsedURL.Host) == "" {
		log.Fatal("EMAIL_CONFIRMATION_BASE_URL must be a valid absolute URL")
	}
}
