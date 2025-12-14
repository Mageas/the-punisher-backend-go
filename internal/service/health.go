package service

import (
	"context"

	"github.com/mageas/the-punisher-backend/internal/platform/config"
)

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
)

type Pinger interface {
	Ping(context.Context) error
}

type HealthService struct {
	config *config.Config
	db     Pinger
}

func NewHealthService(cfg *config.Config, db Pinger) *HealthService {
	return &HealthService{
		config: cfg,
		db:     db,
	}
}

type HealthCheck struct {
	Status      string            `json:"status"`
	Environment string            `json:"environment"`
	Version     string            `json:"version"`
	Services    map[string]string `json:"services"`
}

func (s *HealthService) Check() HealthCheck {
	check := HealthCheck{
		Status:      StatusHealthy,
		Environment: s.config.Env,
		Version:     s.config.Version,
		Services: map[string]string{
			"database": StatusHealthy,
		},
	}

	if err := s.db.Ping(context.Background()); err != nil {
		check.Status = StatusUnhealthy
		check.Services["database"] = StatusUnhealthy
	}

	return check
}
