package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type DashboardService interface {
	GetDashboard(ctx context.Context, userID uuid.UUID, classroomID *uuid.UUID) (*dto.ReturnDashboardDto, error)
}

type dashboardService struct {
	repo repository.Querier
}

func NewDashboardService(repo repository.Querier) DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetDashboard(ctx context.Context, userID uuid.UUID, classroomID *uuid.UUID) (*dto.ReturnDashboardDto, error) {
	classroomIDParam := classroomID
	if classroomID != nil {
		if _, err := s.repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
			ID:     *classroomID,
			UserID: userID,
		}); err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return nil, api.ErrClassroomNotFound
			}
			return nil, fmt.Errorf("failed to get classroom: %w", err)
		}

	}

	kpis, err := s.repo.GetDashboardKpis(ctx, repository.GetDashboardKpisParams{
		UserID:      userID,
		ClassroomID: classroomIDParam,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard kpis: %w", err)
	}

	return sqlcmapper.DashboardFromRow(&kpis), nil
}
