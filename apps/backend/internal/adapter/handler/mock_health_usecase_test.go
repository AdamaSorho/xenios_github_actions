package handler

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockGetHealthStatusUseCase is a mock implementation of the GetHealthStatus use case for testing.
type mockGetHealthStatusUseCase struct {
	executeFunc func(ctx context.Context) (*entities.Health, error)
}

func (m *mockGetHealthStatusUseCase) Execute(ctx context.Context) (*entities.Health, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx)
	}
	return &entities.Health{
		Status: "healthy",
		Checks: map[string]entities.HealthCheck{
			"database": {Status: "up", LatencyMs: 10},
		},
		Uptime: "1h30m",
	}, nil
}
