package config

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

func TestGetEnvHelpers(t *testing.T) {
	t.Setenv("KEY_STR", "value")
	t.Setenv("KEY_BOOL", "true")
	t.Setenv("KEY_DUR", "7")
	t.Setenv("KEY_INT", "42")
	t.Setenv("KEY_CSV", "a, b, ,c")

	if got := GetEnv("KEY_STR", "fallback"); got != "value" {
		t.Fatalf("GetEnv mismatch: %s", got)
	}
	if got := GetEnvBool("KEY_BOOL", false); !got {
		t.Fatalf("GetEnvBool mismatch")
	}
	if got := GetEnvDuration("KEY_DUR", 1); got != 7 {
		t.Fatalf("GetEnvDuration mismatch: %v", got)
	}
	if got := GetEnvInt("KEY_INT", 1); got != 42 {
		t.Fatalf("GetEnvInt mismatch: %d", got)
	}

	if got := GetEnvCSV("KEY_CSV", []string{"x"}); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("GetEnvCSV mismatch: %#v", got)
	}
}

func TestGetEnvHelpersFallback(t *testing.T) {
	if got := GetEnv("MISSING", "fallback"); got != "fallback" {
		t.Fatalf("GetEnv fallback mismatch: %s", got)
	}
	if got := GetEnvBool("MISSING_BOOL", true); !got {
		t.Fatalf("GetEnvBool fallback mismatch")
	}
	if got := GetEnvDuration("MISSING_DUR", 3); got != 3 {
		t.Fatalf("GetEnvDuration fallback mismatch: %v", got)
	}
	if got := GetEnvInt("MISSING_INT", 9); got != 9 {
		t.Fatalf("GetEnvInt fallback mismatch: %d", got)
	}
	if got := GetEnvCSV("MISSING_CSV", []string{"x"}); !reflect.DeepEqual(got, []string{"x"}) {
		t.Fatalf("GetEnvCSV fallback mismatch: %#v", got)
	}
}

func TestLoad(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_ADDRESS", ":9090")
	t.Setenv("APP_VERSION", "9.9.9")
	t.Setenv("APP_ALLOW_REGISTER", "false")
	t.Setenv("APP_DATABASE_URL", "postgres://localhost/test")

	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	t.Setenv("CORS_ALLOWED_METHODS", "GET,POST")
	t.Setenv("CORS_ALLOWED_HEADERS", "Accept,Content-Type")
	t.Setenv("CORS_EXPOSED_HEADERS", "Link")
	t.Setenv("CORS_ALLOW_CREDENTIALS", "true")
	t.Setenv("CORS_MAX_AGE", "600")

	t.Setenv("JWT_ACCESS_SECRET", "access")
	t.Setenv("JWT_ACCESS_EXPIRATION_IN_MINUTES", "15")
	t.Setenv("JWT_REFRESH_SECRET", "refresh")
	t.Setenv("JWT_REFRESH_EXPIRATION_IN_DAYS", "7")
	t.Setenv("JWT_REFRESH_COOKIE_SECURE", "true")
	t.Setenv("JWT_ISSUER", "issuer")
	t.Setenv("JWT_AUDIENCE", "aud")
	t.Setenv("EMAIL_CONFIRMATION_SECRET", "email-confirm-secret")
	t.Setenv("EMAIL_CONFIRMATION_EXPIRATION_IN_HOURS", "48")
	t.Setenv("EMAIL_CONFIRMATION_BASE_URL", "http://localhost:8080/v1/auth/confirm-email")

	t.Setenv("SMTP_HOST", "localhost")
	t.Setenv("SMTP_PORT", "1025")
	t.Setenv("SMTP_USERNAME", "smtp-user")
	t.Setenv("SMTP_PASSWORD", "smtp-password")
	t.Setenv("SMTP_FROM_EMAIL", "no-reply@test.local")
	t.Setenv("SMTP_FROM_NAME", "Punisher Bot")

	cfg := Load()
	if cfg.Env != "test" || cfg.Addr != ":9090" || cfg.Version != "9.9.9" {
		t.Fatalf("unexpected base config: %+v", cfg)
	}
	if cfg.AllowRegister {
		t.Fatalf("expected allow register false")
	}
	if cfg.CORS.MaxAge != 600 {
		t.Fatalf("unexpected CORS max age: %d", cfg.CORS.MaxAge)
	}
	if cfg.JWT.Issuer != "issuer" || cfg.JWT.Audience != "aud" {
		t.Fatalf("unexpected jwt config: %+v", cfg.JWT)
	}
	if cfg.EmailConfirm.Secret != "email-confirm-secret" || cfg.EmailConfirm.Expiration.Hours() != 48 {
		t.Fatalf("unexpected email confirmation config: %+v", cfg.EmailConfirm)
	}
	if cfg.SMTP.Host != "localhost" || cfg.SMTP.Port != 1025 || cfg.SMTP.FromEmail != "no-reply@test.local" {
		t.Fatalf("unexpected smtp config: %+v", cfg.SMTP)
	}
}

func TestValidateCORSConfigValid(t *testing.T) {
	validateCORSConfig(CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

func TestGetEnvOrFatal(t *testing.T) {
	t.Setenv("PRESENT", "ok")
	if got := GetEnvOrFatal("PRESENT"); got != "ok" {
		t.Fatalf("unexpected value: %s", got)
	}
}

func runFatalHelper(t *testing.T, helperCase string) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessForFatalConfig")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "HELPER_CASE="+helperCase)

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected helper process to fail for %s", helperCase)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() == 0 {
		t.Fatalf("expected non-zero exit code")
	}
}

func TestGetEnvOrFatalMissingExits(t *testing.T) {
	runFatalHelper(t, "getenvorfatal")
}

func TestValidateCORSConfigNoOriginsExits(t *testing.T) {
	runFatalHelper(t, "cors_no_origins")
}

func TestValidateCORSConfigNegativeMaxAgeExits(t *testing.T) {
	runFatalHelper(t, "cors_negative_maxage")
}

func TestValidateCORSConfigWildcardWithCredentialsExits(t *testing.T) {
	runFatalHelper(t, "cors_wildcard_credentials")
}

func TestValidateSMTPConfigEmptyHostExits(t *testing.T) {
	runFatalHelper(t, "smtp_empty_host")
}

func TestValidateSMTPConfigInvalidPortExits(t *testing.T) {
	runFatalHelper(t, "smtp_invalid_port")
}

func TestValidateEmailConfirmationConfigInvalidDurationExits(t *testing.T) {
	runFatalHelper(t, "email_confirmation_invalid_duration")
}

func TestValidateEmailConfirmationConfigInvalidBaseURLExits(t *testing.T) {
	runFatalHelper(t, "email_confirmation_invalid_base_url")
}

func TestHelperProcessForFatalConfig(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	switch os.Getenv("HELPER_CASE") {
	case "getenvorfatal":
		_ = GetEnvOrFatal("MISSING_KEY")
	case "cors_no_origins":
		validateCORSConfig(CORSConfig{AllowedOrigins: []string{}, MaxAge: 1})
	case "cors_negative_maxage":
		validateCORSConfig(CORSConfig{AllowedOrigins: []string{"http://localhost"}, MaxAge: -1})
	case "cors_wildcard_credentials":
		validateCORSConfig(CORSConfig{AllowedOrigins: []string{"*"}, AllowCredentials: true, MaxAge: 1})
	case "smtp_empty_host":
		validateSMTPConfig(SMTPConfig{Host: "", Port: 1025, FromEmail: "no-reply@test.local"})
	case "smtp_invalid_port":
		validateSMTPConfig(SMTPConfig{Host: "localhost", Port: 0, FromEmail: "no-reply@test.local"})
	case "email_confirmation_invalid_duration":
		validateEmailConfirmationConfig(EmailConfirmationConfig{
			Secret:     "secret",
			Expiration: 0,
			BaseURL:    "http://localhost:8080/v1/auth/confirm-email",
		})
	case "email_confirmation_invalid_base_url":
		validateEmailConfirmationConfig(EmailConfirmationConfig{
			Secret:     "secret",
			Expiration: 24 * time.Hour,
			BaseURL:    "/v1/auth/confirm-email",
		})
	default:
		os.Exit(2)
	}

	os.Exit(0)
}
