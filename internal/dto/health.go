package dto

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
)

type HealthCheckDto struct {
	Status      string            `json:"status"`
	Environment string            `json:"environment"`
	Version     string            `json:"version"`
	Services    map[string]string `json:"services"`
}
