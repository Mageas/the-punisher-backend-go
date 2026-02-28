package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestHealthServiceHealthy(t *testing.T) {
	svc := NewHealthService(&config.Config{Env: "test", Version: "v1"}, fakePinger{})
	check := svc.Check()
	if check.Status != dto.StatusHealthy {
		t.Fatalf("expected healthy, got %s", check.Status)
	}
	if check.Services["database"] != dto.StatusHealthy {
		t.Fatalf("expected db healthy")
	}
}

func TestHealthServiceUnhealthy(t *testing.T) {
	svc := NewHealthService(&config.Config{Env: "test", Version: "v1"}, fakePinger{err: errors.New("db down")})
	check := svc.Check()
	if check.Status != dto.StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", check.Status)
	}
	if check.Services["database"] != dto.StatusUnhealthy {
		t.Fatalf("expected db unhealthy")
	}
}
