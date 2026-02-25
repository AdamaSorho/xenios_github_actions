package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_CreateBatch_StoresMeasurements(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "system",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           85.4,
			Unit:            "kg",
			MeasuredAt:      time.Now(),
			ArtifactID:      "artifact-1",
		},
		{
			ClientID:        "client-1",
			RecordedBy:      "system",
			MeasurementType: entities.MeasurementTypeBodyFatPct,
			Value:           22.3,
			Unit:            "%",
			MeasuredAt:      time.Now(),
			ArtifactID:      "artifact-1",
		},
	}

	results, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.ID == "" {
			t.Error("expected non-empty ID")
		}
	}
}

func TestInMemoryMeasurementRepository_CreateBatch_GeneratesIDs(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "system",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           85.4,
			Unit:            "kg",
			MeasuredAt:      time.Now(),
		},
	}

	results, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].ID == "" {
		t.Error("expected generated ID")
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_ReturnsMatching(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "system",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           85.4,
			Unit:            "kg",
			MeasuredAt:      time.Now(),
			ArtifactID:      "artifact-1",
		},
		{
			ClientID:        "client-1",
			RecordedBy:      "system",
			MeasurementType: entities.MeasurementTypeBMR,
			Value:           1847,
			Unit:            "kcal",
			MeasuredAt:      time.Now(),
			ArtifactID:      "artifact-2",
		},
	}

	_, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := repo.FindByArtifactID(context.Background(), "artifact-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 measurement, got %d", len(found))
	}
	if found[0].MeasurementType != entities.MeasurementTypeWeight {
		t.Errorf("expected weight measurement, got %s", found[0].MeasurementType)
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_ReturnsEmptyForUnknown(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	found, err := repo.FindByArtifactID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(found))
	}
}

func TestInMemoryFileDownloader_Download_ReturnsContent(t *testing.T) {
	downloader := NewInMemoryFileDownloader()
	downloader.PutFile("test-key", []byte("pdf content"))

	content, err := downloader.Download(context.Background(), "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(content) != "pdf content" {
		t.Errorf("expected 'pdf content', got %q", string(content))
	}
}

func TestInMemoryFileDownloader_Download_ReturnsErrorForMissing(t *testing.T) {
	downloader := NewInMemoryFileDownloader()

	_, err := downloader.Download(context.Background(), "missing-key")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
