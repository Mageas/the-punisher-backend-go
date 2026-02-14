package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mageas/the-punisher-backend/internal/api/handler"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
)

type staticHealthChecker struct {
	data dto.HealthCheckDto
}

func (s staticHealthChecker) Check() dto.HealthCheckDto {
	return s.data
}

func TestHealthHandlerGetHealth(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		h := handler.NewHealthHandler(staticHealthChecker{
			data: dto.HealthCheckDto{
				Status:      dto.StatusHealthy,
				Environment: "test",
				Version:     "1.0.0",
				Services: map[string]string{
					"database": dto.StatusHealthy,
				},
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
		rr := httptest.NewRecorder()
		h.GetHealth(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[dto.HealthCheckDto](t, rr)
		if resp.Status != dto.StatusHealthy {
			t.Fatalf("expected status %q, got %q", dto.StatusHealthy, resp.Status)
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		h := handler.NewHealthHandler(staticHealthChecker{
			data: dto.HealthCheckDto{
				Status:      dto.StatusUnhealthy,
				Environment: "test",
				Version:     "1.0.0",
				Services: map[string]string{
					"database": dto.StatusUnhealthy,
				},
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
		rr := httptest.NewRecorder()
		h.GetHealth(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[dto.HealthCheckDto](t, rr)
		if resp.Status != dto.StatusUnhealthy {
			t.Fatalf("expected status %q, got %q", dto.StatusUnhealthy, resp.Status)
		}
	})
}
