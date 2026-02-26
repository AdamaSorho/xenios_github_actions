package entities

import (
	"testing"
	"time"
)

func TestIsValidMeasurementType_KnownTypes_ReturnsTrue(t *testing.T) {
	types := []MeasurementType{
		MeasurementTypeWeight,
		MeasurementTypeBodyFatPct,
		MeasurementTypeSkeletalMuscleMass,
		MeasurementTypeBMR,
		MeasurementTypeTotalBodyWater,
		MeasurementTypeLeanBodyMass,
	}
	for _, mt := range types {
		if !IsValidMeasurementType(mt) {
			t.Errorf("expected %q to be valid", mt)
		}
	}
}

func TestIsValidMeasurementType_UnknownType_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementType("unknown_type") {
		t.Error("expected unknown type to be invalid")
	}
}

func TestIsValidMeasurementType_Empty_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementType("") {
		t.Error("expected empty type to be invalid")
	}
}

func validMeasurement() *Measurement {
	return &Measurement{
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: MeasurementTypeWeight,
		Value:           85.4,
		Unit:            "kg",
		MeasuredAt:      time.Now(),
	}
}

func TestValidateNewMeasurement_Valid_ReturnsNil(t *testing.T) {
	m := validMeasurement()
	if err := ValidateNewMeasurement(m); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateNewMeasurement_MissingClientID_ReturnsError(t *testing.T) {
	m := validMeasurement()
	m.ClientID = ""
	err := ValidateNewMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if ve.Message != "client_id is required" {
		t.Errorf("unexpected message: %s", ve.Message)
	}
}

func TestValidateNewMeasurement_MissingRecordedBy_ReturnsError(t *testing.T) {
	m := validMeasurement()
	m.RecordedBy = ""
	err := ValidateNewMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateNewMeasurement_InvalidType_ReturnsError(t *testing.T) {
	m := validMeasurement()
	m.MeasurementType = "invalid"
	err := ValidateNewMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateNewMeasurement_MissingUnit_ReturnsError(t *testing.T) {
	m := validMeasurement()
	m.Unit = ""
	err := ValidateNewMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateNewMeasurement_ZeroMeasuredAt_ReturnsError(t *testing.T) {
	m := validMeasurement()
	m.MeasuredAt = time.Time{}
	err := ValidateNewMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractionResult_IsPartial_DefaultsFalse(t *testing.T) {
	result := &ExtractionResult{}
	if result.IsPartial {
		t.Error("expected IsPartial to be false by default")
	}
}

func TestExtractionResult_WithMeasurements(t *testing.T) {
	result := &ExtractionResult{
		Measurements: []*Measurement{validMeasurement()},
		IsPartial:    false,
	}
	if len(result.Measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result.Measurements))
	}
}
