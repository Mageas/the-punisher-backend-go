package service

import (
	"context"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
)

type HealthChecker interface {
	Check() dto.HealthCheckDto
}

type Pinger interface {
	Ping(context.Context) error
}

type healthService struct {
	config *config.Config
	db     Pinger
}

func NewHealthService(cfg *config.Config, db Pinger) HealthChecker {
	return &healthService{
		config: cfg,
		db:     db,
	}
}

func (s *healthService) Check() dto.HealthCheckDto {
	check := dto.HealthCheckDto{
		Status:      dto.StatusHealthy,
		Environment: s.config.Env,
		Version:     s.config.Version,
		Services: map[string]string{
			"database": dto.StatusHealthy,
		},
	}

	if err := s.db.Ping(context.Background()); err != nil {
		check.Status = dto.StatusUnhealthy
		check.Services["database"] = dto.StatusUnhealthy
	}

	return check
}
