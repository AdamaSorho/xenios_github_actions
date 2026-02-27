package entities

import "testing"

func TestMeasurement_IsOutOfRange_FlaggedHigh_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagHigh}
	if !m.IsOutOfRange() {
		t.Error("expected high flag to be out of range")
	}
}

func TestMeasurement_IsOutOfRange_FlaggedLow_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagLow}
	if !m.IsOutOfRange() {
		t.Error("expected low flag to be out of range")
	}
}

func TestMeasurement_IsOutOfRange_CriticalHigh_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagCriticalHigh}
	if !m.IsOutOfRange() {
		t.Error("expected critical_high flag to be out of range")
	}
}

func TestMeasurement_IsOutOfRange_CriticalLow_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagCriticalLow}
	if !m.IsOutOfRange() {
		t.Error("expected critical_low flag to be out of range")
	}
}

func TestMeasurement_IsOutOfRange_Normal_ReturnsFalse(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagNormal}
	if m.IsOutOfRange() {
		t.Error("expected normal flag to not be out of range")
	}
}

func TestMeasurement_IsCritical_CriticalHigh_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagCriticalHigh}
	if !m.IsCritical() {
		t.Error("expected critical_high to be critical")
	}
}

func TestMeasurement_IsCritical_CriticalLow_ReturnsTrue(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagCriticalLow}
	if !m.IsCritical() {
		t.Error("expected critical_low to be critical")
	}
}

func TestMeasurement_IsCritical_High_ReturnsFalse(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagHigh}
	if m.IsCritical() {
		t.Error("expected high flag to not be critical")
	}
}

func TestMeasurement_IsCritical_Normal_ReturnsFalse(t *testing.T) {
	m := &Measurement{Flag: MeasurementFlagNormal}
	if m.IsCritical() {
		t.Error("expected normal flag to not be critical")
	}
}
