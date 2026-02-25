package entities

import "testing"

func TestIsValidMeasurementType_KnownTypes_ReturnsTrue(t *testing.T) {
	types := []MeasurementType{
		MeasurementTypeCalories,
		MeasurementTypeProtein,
		MeasurementTypeCarbs,
		MeasurementTypeFat,
		MeasurementTypeFiber,
	}
	for _, mt := range types {
		t.Run(string(mt), func(t *testing.T) {
			if !IsValidMeasurementType(mt) {
				t.Errorf("expected %s to be valid", mt)
			}
		})
	}
}

func TestIsValidMeasurementType_UnknownType_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementType("unknown_type") {
		t.Error("expected unknown_type to be invalid")
	}
}

func TestMeasurementUnit_Calories_ReturnsKcal(t *testing.T) {
	unit := MeasurementUnit(MeasurementTypeCalories)
	if unit != "kcal" {
		t.Errorf("expected 'kcal', got '%s'", unit)
	}
}

func TestMeasurementUnit_Macros_ReturnsGrams(t *testing.T) {
	macros := []MeasurementType{
		MeasurementTypeProtein,
		MeasurementTypeCarbs,
		MeasurementTypeFat,
		MeasurementTypeFiber,
	}
	for _, mt := range macros {
		t.Run(string(mt), func(t *testing.T) {
			unit := MeasurementUnit(mt)
			if unit != "g" {
				t.Errorf("expected 'g', got '%s'", unit)
			}
		})
	}
}

func TestAllMeasurementTypes_ReturnsFiveTypes(t *testing.T) {
	types := AllMeasurementTypes()
	if len(types) != 5 {
		t.Errorf("expected 5 types, got %d", len(types))
	}
}

func TestIsValidJobType_ExtractNutrition_ReturnsTrue(t *testing.T) {
	if !IsValidJobType(JobTypeExtractNutrition) {
		t.Error("expected extract_nutrition to be a valid job type")
	}
}
