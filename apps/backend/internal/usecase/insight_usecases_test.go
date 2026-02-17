package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// --- Mock InsightCard Repository ---

type mockInsightCardRepo struct {
	findByIDFunc   func(ctx context.Context, id string) (*entities.InsightCard, error)
	listByCoachFn  func(ctx context.Context, filter repository.InsightCardFilter) ([]*entities.InsightCard, int, error)
	listByClientFn func(ctx context.Context, filter repository.InsightCardFilter) ([]*entities.InsightCard, int, error)
	updateFunc     func(ctx context.Context, insight *entities.InsightCard) error
	createFunc     func(ctx context.Context, insight *entities.InsightCard) error
}

func (m *mockInsightCardRepo) FindByID(ctx context.Context, id string) (*entities.InsightCard, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockInsightCardRepo) ListByCoach(ctx context.Context, filter repository.InsightCardFilter) ([]*entities.InsightCard, int, error) {
	if m.listByCoachFn != nil {
		return m.listByCoachFn(ctx, filter)
	}
	return []*entities.InsightCard{}, 0, nil
}

func (m *mockInsightCardRepo) ListByClient(ctx context.Context, filter repository.InsightCardFilter) ([]*entities.InsightCard, int, error) {
	if m.listByClientFn != nil {
		return m.listByClientFn(ctx, filter)
	}
	return []*entities.InsightCard{}, 0, nil
}

func (m *mockInsightCardRepo) Update(ctx context.Context, insight *entities.InsightCard) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, insight)
	}
	return nil
}

func (m *mockInsightCardRepo) Create(ctx context.Context, insight *entities.InsightCard) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, insight)
	}
	return nil
}

// --- Mock Audit Repository ---

type mockAuditRepo struct {
	loggedEvents []*entities.AuditEvent
}

func (m *mockAuditRepo) LogEvent(_ context.Context, event *entities.AuditEvent) error {
	m.loggedEvents = append(m.loggedEvents, event)
	return nil
}

func (m *mockAuditRepo) Query(_ context.Context, _ entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	return nil, 0, nil
}

// --- GetInsightQueueUseCase Tests ---

func TestGetInsightQueue_Success(t *testing.T) {
	now := time.Now()
	insights := []*entities.InsightCard{
		{ID: "1", CoachID: "coach-1", ClientID: "client-1", Status: "draft", Priority: "high", CreatedAt: now},
		{ID: "2", CoachID: "coach-1", ClientID: "client-2", Status: "draft", Priority: "medium", CreatedAt: now},
	}
	repo := &mockInsightCardRepo{
		listByCoachFn: func(_ context.Context, f repository.InsightCardFilter) ([]*entities.InsightCard, int, error) {
			if f.CoachID != "coach-1" {
				t.Errorf("expected coach-1, got %s", f.CoachID)
			}
			if f.Status != "draft" {
				t.Errorf("expected draft filter, got %s", f.Status)
			}
			return insights, 2, nil
		},
	}
	uc := NewGetInsightQueueUseCase(repo)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{CoachID: "coach-1", Status: "draft", Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 2 {
		t.Errorf("expected 2 insights, got %d", len(out.Insights))
	}
	if out.Total != 2 {
		t.Errorf("expected total 2, got %d", out.Total)
	}
}

func TestGetInsightQueue_EmptyCoachID(t *testing.T) {
	repo := &mockInsightCardRepo{}
	uc := NewGetInsightQueueUseCase(repo)

	_, err := uc.Execute(context.Background(), GetInsightQueueInput{CoachID: ""})
	if err == nil {
		t.Fatal("expected error for empty coach ID")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetInsightQueue_DefaultLimit(t *testing.T) {
	repo := &mockInsightCardRepo{
		listByCoachFn: func(_ context.Context, f repository.InsightCardFilter) ([]*entities.InsightCard, int, error) {
			if f.Limit != 20 {
				t.Errorf("expected default limit 20, got %d", f.Limit)
			}
			return []*entities.InsightCard{}, 0, nil
		},
	}
	uc := NewGetInsightQueueUseCase(repo)
	_, _ = uc.Execute(context.Background(), GetInsightQueueInput{CoachID: "coach-1"})
}

// --- ApproveInsightUseCase Tests ---

func TestApproveInsight_Success(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", ClientID: "client-1", Status: "draft"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, id string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	audit := &mockAuditRepo{}
	uc := NewApproveInsightUseCase(repo, audit)

	out, err := uc.Execute(context.Background(), "i-1", "coach-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Status != "approved" {
		t.Errorf("expected approved, got %s", out.Status)
	}
	if len(audit.loggedEvents) != 1 {
		t.Errorf("expected 1 audit event, got %d", len(audit.loggedEvents))
	}
	if audit.loggedEvents[0].Action != "insight.approve" {
		t.Errorf("expected insight.approve action, got %s", audit.loggedEvents[0].Action)
	}
}

func TestApproveInsight_NotFound(t *testing.T) {
	repo := &mockInsightCardRepo{}
	uc := NewApproveInsightUseCase(repo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "nonexistent", "coach-1")
	if err == nil {
		t.Fatal("expected error for nonexistent insight")
	}
}

func TestApproveInsight_WrongCoach(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "draft"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	uc := NewApproveInsightUseCase(repo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "i-1", "coach-2")
	if err == nil {
		t.Fatal("expected error for wrong coach")
	}
}

func TestApproveInsight_AlreadyDismissed(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "dismissed"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	uc := NewApproveInsightUseCase(repo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "i-1", "coach-1")
	if err == nil {
		t.Fatal("expected error for already dismissed insight")
	}
}

// --- DismissInsightUseCase Tests ---

func TestDismissInsight_Success(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "draft"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	audit := &mockAuditRepo{}
	uc := NewDismissInsightUseCase(repo, audit)

	out, err := uc.Execute(context.Background(), "i-1", "coach-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Status != "dismissed" {
		t.Errorf("expected dismissed, got %s", out.Status)
	}
	if len(audit.loggedEvents) != 1 {
		t.Errorf("expected 1 audit event, got %d", len(audit.loggedEvents))
	}
}

func TestDismissInsight_WrongCoach(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "draft"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	uc := NewDismissInsightUseCase(repo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "i-1", "coach-2")
	if err == nil {
		t.Fatal("expected error for wrong coach")
	}
}

// --- EditInsightUseCase Tests ---

func TestEditInsight_Success(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "draft", Title: "Old", Body: "Old body"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	audit := &mockAuditRepo{}
	uc := NewEditInsightUseCase(repo, audit)

	out, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "i-1",
		CoachID:   "coach-1",
		Title:     "New Title",
		Body:      "New Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Title != "New Title" {
		t.Errorf("expected 'New Title', got %s", out.Title)
	}
	if out.Body != "New Body" {
		t.Errorf("expected 'New Body', got %s", out.Body)
	}
	if len(audit.loggedEvents) != 1 {
		t.Errorf("expected 1 audit event, got %d", len(audit.loggedEvents))
	}
}

func TestEditInsight_EmptyTitle(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "draft"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	uc := NewEditInsightUseCase(repo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "i-1",
		CoachID:   "coach-1",
		Title:     "",
		Body:      "body",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestEditInsight_WrongCoach(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "draft"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	uc := NewEditInsightUseCase(repo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "i-1",
		CoachID:   "coach-2",
		Title:     "title",
		Body:      "body",
	})
	if err == nil {
		t.Fatal("expected error for wrong coach")
	}
}

// --- ShareInsightUseCase Tests ---

func TestShareInsight_Success(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "approved"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	audit := &mockAuditRepo{}
	uc := NewShareInsightUseCase(repo, audit)

	out, err := uc.Execute(context.Background(), "i-1", "coach-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Status != "shared" {
		t.Errorf("expected shared, got %s", out.Status)
	}
	if len(audit.loggedEvents) != 1 {
		t.Errorf("expected 1 audit event, got %d", len(audit.loggedEvents))
	}
}

func TestShareInsight_FromDraft_Fails(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "draft"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	uc := NewShareInsightUseCase(repo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "i-1", "coach-1")
	if err == nil {
		t.Fatal("expected error when sharing draft")
	}
}

func TestShareInsight_WrongCoach(t *testing.T) {
	insight := &entities.InsightCard{ID: "i-1", CoachID: "coach-1", Status: "approved"}
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return insight, nil
		},
	}
	uc := NewShareInsightUseCase(repo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "i-1", "coach-2")
	if err == nil {
		t.Fatal("expected error for wrong coach")
	}
}

// --- GetClientInsightsUseCase Tests ---

func TestGetClientInsights_Success(t *testing.T) {
	now := time.Now()
	insights := []*entities.InsightCard{
		{ID: "1", ClientID: "client-1", Status: "approved", CreatedAt: now},
	}
	repo := &mockInsightCardRepo{
		listByClientFn: func(_ context.Context, f repository.InsightCardFilter) ([]*entities.InsightCard, int, error) {
			return insights, 1, nil
		},
	}
	uc := NewGetClientInsightsUseCase(repo)

	out, err := uc.Execute(context.Background(), GetClientInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 1 {
		t.Errorf("expected 1 insight, got %d", len(out.Insights))
	}
}

func TestGetClientInsights_EmptyClientID(t *testing.T) {
	repo := &mockInsightCardRepo{}
	uc := NewGetClientInsightsUseCase(repo)

	_, err := uc.Execute(context.Background(), GetClientInsightsInput{ClientID: ""})
	if err == nil {
		t.Fatal("expected error for empty client ID")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestApproveInsight_RepoError(t *testing.T) {
	repo := &mockInsightCardRepo{
		findByIDFunc: func(_ context.Context, _ string) (*entities.InsightCard, error) {
			return nil, errors.New("db error")
		},
	}
	uc := NewApproveInsightUseCase(repo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "i-1", "coach-1")
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
