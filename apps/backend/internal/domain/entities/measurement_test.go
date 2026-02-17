package entities

import "testing"

func TestMeasurement_IsOutOfRange_HighFlag_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagHigh}
	if !m.IsOutOfRange() {
		t.Error("expected IsOutOfRange to return true for high flag")
	}
}

func TestMeasurement_IsOutOfRange_LowFlag_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagLow}
	if !m.IsOutOfRange() {
		t.Error("expected IsOutOfRange to return true for low flag")
	}
}

func TestMeasurement_IsOutOfRange_NormalFlag_ReturnsFalse(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagNormal}
	if m.IsOutOfRange() {
		t.Error("expected IsOutOfRange to return false for normal flag")
	}
}

func TestMeasurement_IsOutOfRange_CriticalFlag_ReturnsFalse(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagCriticalHigh}
	if m.IsOutOfRange() {
		t.Error("expected IsOutOfRange to return false for critical_high flag (critical is separate)")
	}
}

func TestMeasurement_IsCritical_CriticalHighFlag_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagCriticalHigh}
	if !m.IsCritical() {
		t.Error("expected IsCritical to return true for critical_high flag")
	}
}

func TestMeasurement_IsCritical_CriticalLowFlag_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagCriticalLow}
	if !m.IsCritical() {
		t.Error("expected IsCritical to return true for critical_low flag")
	}
}

func TestMeasurement_IsCritical_NormalFlag_ReturnsFalse(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagNormal}
	if m.IsCritical() {
		t.Error("expected IsCritical to return false for normal flag")
	}
}

func TestMeasurement_IsCritical_HighFlag_ReturnsFalse(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagHigh}
	if m.IsCritical() {
		t.Error("expected IsCritical to return false for high flag (only critical flags)")
	}
}
