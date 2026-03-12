package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const defaultUserTimezone = "Europe/Paris"

func resolveUserLocation(ctx context.Context, repo repository.Querier, userID uuid.UUID) (*time.Location, error) {
	user, err := repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to get user timezone: %w", err)
	}

	timezone := strings.TrimSpace(user.Timezone)
	if timezone == "" {
		timezone = defaultUserTimezone
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load user timezone %q: %w", timezone, err)
	}

	return location, nil
}

func calendarDateInLocation(value time.Time, location *time.Location) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, location)
}

func startOfDayForInstant(value time.Time, location *time.Location) time.Time {
	localValue := value.In(location)
	return time.Date(localValue.Year(), localValue.Month(), localValue.Day(), 0, 0, 0, 0, location)
}

func localDateBoundsToUTC(from, to *time.Time, location *time.Location) (*time.Time, *time.Time) {
	var fromUTC *time.Time
	if from != nil {
		start := calendarDateInLocation(*from, location).UTC()
		fromUTC = &start
	}

	var toUTC *time.Time
	if to != nil {
		endExclusive := calendarDateInLocation(*to, location).AddDate(0, 0, 1).UTC()
		toUTC = &endExclusive
	}

	return fromUTC, toUTC
}
