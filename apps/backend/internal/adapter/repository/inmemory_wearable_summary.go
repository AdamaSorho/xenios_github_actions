package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryWearableSummaryRepository is an in-memory implementation of WearableSummaryRepository.
type InMemoryWearableSummaryRepository struct {
	mu      sync.RWMutex
	records []*entities.WearableSummary
}

// NewInMemoryWearableSummaryRepository creates a new in-memory wearable summary repository.
func NewInMemoryWearableSummaryRepository() *InMemoryWearableSummaryRepository {
	return &InMemoryWearableSummaryRepository{
		records: make([]*entities.WearableSummary, 0),
	}
}

// Upsert creates or updates a wearable summary.
func (r *InMemoryWearableSummaryRepository) Upsert(_ context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for existing record with same client, source, and date
	for i, rec := range r.records {
		if rec.ClientID == summary.ClientID &&
			rec.Source == summary.Source &&
			rec.SummaryDate.Equal(summary.SummaryDate) {
			// Update existing
			r.records[i] = &entities.WearableSummary{
				ID:          rec.ID,
				ClientID:    summary.ClientID,
				Source:      summary.Source,
				SummaryDate: summary.SummaryDate,
				Metrics:     summary.Metrics,
				SyncedAt:    summary.SyncedAt,
				CreatedAt:   rec.CreatedAt,
			}
			return r.records[i], nil
		}
	}

	// Create new
	record := &entities.WearableSummary{
		ID:          generateRepoID(),
		ClientID:    summary.ClientID,
		Source:      summary.Source,
		SummaryDate: summary.SummaryDate,
		Metrics:     summary.Metrics,
		SyncedAt:    summary.SyncedAt,
		CreatedAt:   time.Now(),
	}
	r.records = append(r.records, record)
	return record, nil
}

// FindByClientID retrieves wearable summaries for a client with filtering.
func (r *InMemoryWearableSummaryRepository) FindByClientID(_ context.Context, filter entities.WearableSummaryFilter) ([]*entities.WearableSummary, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.WearableSummary
	for _, rec := range r.records {
		if rec.ClientID != filter.ClientID {
			continue
		}
		if filter.Source != "" && rec.Source != filter.Source {
			continue
		}
		if filter.From != nil && rec.SummaryDate.Before(*filter.From) {
			continue
		}
		if filter.To != nil && rec.SummaryDate.After(*filter.To) {
			continue
		}
		filtered = append(filtered, rec)
	}

	// Sort by summary_date descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].SummaryDate.After(filtered[j].SummaryDate)
	})

	total := len(filtered)

	if filter.Offset >= total {
		return []*entities.WearableSummary{}, total, nil
	}

	end := filter.Offset + filter.Limit
	if end > total || filter.Limit == 0 {
		end = total
	}

	return filtered[filter.Offset:end], total, nil
}

// FindAverages computes rolling averages from wearable data for the last N days.
func (r *InMemoryWearableSummaryRepository) FindAverages(_ context.Context, clientID string, days int) (*entities.WearableAverages, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cutoff := time.Now().AddDate(0, 0, -days)
	var source string
	var hrvSum, sleepSum, recoverySum float64
	var hrvCount, sleepCount, recoveryCount int

	for _, rec := range r.records {
		if rec.ClientID != clientID || rec.SummaryDate.Before(cutoff) {
			continue
		}
		if source == "" {
			source = rec.Source
		}
		if v, ok := rec.Metrics["hrv"].(float64); ok {
			hrvSum += v
			hrvCount++
		}
		if v, ok := rec.Metrics["sleep_hours"].(float64); ok {
			sleepSum += v
			sleepCount++
		}
		if v, ok := rec.Metrics["recovery"].(float64); ok {
			recoverySum += v
			recoveryCount++
		}
	}

	if source == "" {
		return nil, nil
	}

	avg := &entities.WearableAverages{Source: source}
	if hrvCount > 0 {
		v := hrvSum / float64(hrvCount)
		avg.AvgHRV7d = &v
	}
	if sleepCount > 0 {
		v := sleepSum / float64(sleepCount)
		avg.AvgSleep7d = &v
	}
	if recoveryCount > 0 {
		v := recoverySum / float64(recoveryCount)
		avg.AvgRecovery7d = &v
	}

	return avg, nil
}
