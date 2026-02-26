package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

type mockMeasurementRepo struct {
	findByClientFunc   func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error)
	findLatestFunc     func(ctx context.Context, clientID string) ([]*entities.Measurement, error)
	createFunc         func(ctx context.Context, m *entities.Measurement) (*entities.Measurement, error)
}

func (m *mockMeasurementRepo) Create(ctx context.Context, meas *entities.Measurement) (*entities.Measurement, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, meas)
	}
	return meas, nil
}

func (m *mockMeasurementRepo) FindByClientID(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
	if m.findByClientFunc != nil {
		return m.findByClientFunc(ctx, filter)
	}
	return []*entities.Measurement{}, 0, nil
}

func (m *mockMeasurementRepo) FindLatestByClientID(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
	if m.findLatestFunc != nil {
		return m.findLatestFunc(ctx, clientID)
	}
	return []*entities.Measurement{}, nil
}

type mockWearableRepo struct {
	upsertFunc     func(ctx context.Context, ws *entities.WearableSummary) (*entities.WearableSummary, error)
	findByClientFunc func(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error)
}

func (m *mockWearableRepo) Upsert(ctx context.Context, ws *entities.WearableSummary) (*entities.WearableSummary, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, ws)
	}
	return ws, nil
}

func (m *mockWearableRepo) FindByClientID(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error) {
	if m.findByClientFunc != nil {
		return m.findByClientFunc(ctx, clientID, limit, offset)
	}
	return []*entities.WearableSummary{}, nil
}

type mockAuditRepo struct {
	events []*entities.AuditEvent
}

func (m *mockAuditRepo) LogEvent(_ context.Context, event *entities.AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockAuditRepo) Query(_ context.Context, _ entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	return m.events, len(m.events), nil
}

func authorizedCCRepo() *mockCoachClientRepo {
	return &mockCoachClientRepo{
		findFunc: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return &entities.CoachClient{ID: "rel-1", CoachID: coachID, ClientID: clientID}, nil
		},
	}
}

func unauthorizedCCRepo() *mockCoachClientRepo {
	return &mockCoachClientRepo{
		findFunc: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, nil
		},
	}
}
