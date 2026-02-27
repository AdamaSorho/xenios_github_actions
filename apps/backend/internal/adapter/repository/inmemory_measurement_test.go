package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

func TestInMemoryMeasurementRepository_CreateBatch_StoresAll(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	flag := entities.LabFlagNormal
	inputs := []domainrepo.MeasurementInput{
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: "fasting_glucose",
			Value:           98,
			Unit:            "mg/dL",
			ReferenceLow:    entities.FloatPtr(70),
			ReferenceHigh:   entities.FloatPtr(100),
			Flag:            &flag,
		},
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: "total_cholesterol",
			Value:           210,
			Unit:            "mg/dL",
			ReferenceHigh:   entities.FloatPtr(200),
		},
	}

	count, err := repo.CreateBatch(ctx, inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	all := repo.All()
	if len(all) != 2 {
		t.Fatalf("stored %d, want 2", len(all))
	}
}

func TestInMemoryMeasurementRepository_CreateBatch_Empty_ReturnsZero(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	count, err := repo.CreateBatch(ctx, []domainrepo.MeasurementInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}
