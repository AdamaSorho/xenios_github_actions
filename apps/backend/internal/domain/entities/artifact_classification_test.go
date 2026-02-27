package entities

import "testing"

func TestClassifyDocument_WithValidHint_ReturnsHint(t *testing.T) {
	result := ClassifyDocument("scan.pdf", "application/pdf", DocumentSubtypeInBodyPDF)
	if result != DocumentSubtypeInBodyPDF {
		t.Errorf("expected %s, got %s", DocumentSubtypeInBodyPDF, result)
	}
}

func TestClassifyDocument_HintOverridesExtension(t *testing.T) {
	result := ClassifyDocument("data.csv", "text/csv", DocumentSubtypeLabCSV)
	if result != DocumentSubtypeLabCSV {
		t.Errorf("expected %s, got %s", DocumentSubtypeLabCSV, result)
	}
}

func TestClassifyDocument_InvalidHint_FallsBackToExtension(t *testing.T) {
	result := ClassifyDocument("recording.mp3", "audio/mpeg", DocumentSubtype("invalid"))
	if result != DocumentSubtypeAudio {
		t.Errorf("expected %s, got %s", DocumentSubtypeAudio, result)
	}
}

func TestClassifyDocument_EmptyHint_FallsBackToExtension(t *testing.T) {
	result := ClassifyDocument("recording.wav", "audio/wav", "")
	if result != DocumentSubtypeAudio {
		t.Errorf("expected %s, got %s", DocumentSubtypeAudio, result)
	}
}

func TestClassifyDocument_AudioByExtension_ReturnsAudio(t *testing.T) {
	tests := []struct {
		fileName    string
		contentType string
	}{
		{"recording.mp3", "audio/mpeg"},
		{"voice.wav", "audio/wav"},
		{"note.aac", "audio/aac"},
	}
	for _, tc := range tests {
		result := ClassifyDocument(tc.fileName, tc.contentType, "")
		if result != DocumentSubtypeAudio {
			t.Errorf("ClassifyDocument(%s, %s) = %s, want %s", tc.fileName, tc.contentType, result, DocumentSubtypeAudio)
		}
	}
}

func TestClassifyDocument_AudioByContentType_ReturnsAudio(t *testing.T) {
	result := ClassifyDocument("unknown", "audio/mpeg", "")
	if result != DocumentSubtypeAudio {
		t.Errorf("expected %s, got %s", DocumentSubtypeAudio, result)
	}
}

func TestClassifyDocument_PDFWithoutHint_ReturnsOther(t *testing.T) {
	result := ClassifyDocument("report.pdf", "application/pdf", "")
	if result != DocumentSubtypeOther {
		t.Errorf("expected %s, got %s", DocumentSubtypeOther, result)
	}
}

func TestClassifyDocument_CSVWithoutHint_ReturnsOther(t *testing.T) {
	result := ClassifyDocument("data.csv", "text/csv", "")
	if result != DocumentSubtypeOther {
		t.Errorf("expected %s, got %s", DocumentSubtypeOther, result)
	}
}

func TestClassifyDocument_JSONWithoutHint_ReturnsOther(t *testing.T) {
	result := ClassifyDocument("export.json", "application/json", "")
	if result != DocumentSubtypeOther {
		t.Errorf("expected %s, got %s", DocumentSubtypeOther, result)
	}
}

func TestClassifyDocument_UnknownExtension_ReturnsOther(t *testing.T) {
	result := ClassifyDocument("data.xyz", "application/octet-stream", "")
	if result != DocumentSubtypeOther {
		t.Errorf("expected %s, got %s", DocumentSubtypeOther, result)
	}
}

func TestIsValidDocumentSubtype_ValidSubtypes_ReturnsTrue(t *testing.T) {
	subtypes := []DocumentSubtype{
		DocumentSubtypeInBodyPDF, DocumentSubtypeLabCSV, DocumentSubtypeLabPDF,
		DocumentSubtypeWearableCSV, DocumentSubtypeWearableJSON,
		DocumentSubtypeNutritionCSV, DocumentSubtypeAudio, DocumentSubtypeOther,
	}
	for _, s := range subtypes {
		if !IsValidDocumentSubtype(s) {
			t.Errorf("expected %s to be valid", s)
		}
	}
}

func TestIsValidDocumentSubtype_InvalidSubtype_ReturnsFalse(t *testing.T) {
	if IsValidDocumentSubtype("unknown_type") {
		t.Error("expected unknown_type to be invalid")
	}
}

func TestDocumentSubtypeToJobType_InBodyPDF_ReturnsExtractInBody(t *testing.T) {
	jt := DocumentSubtypeToJobType(DocumentSubtypeInBodyPDF)
	if jt != JobTypeExtractInBody {
		t.Errorf("expected %s, got %s", JobTypeExtractInBody, jt)
	}
}

func TestDocumentSubtypeToJobType_LabCSV_ReturnsExtractLabResults(t *testing.T) {
	jt := DocumentSubtypeToJobType(DocumentSubtypeLabCSV)
	if jt != JobTypeExtractLabResults {
		t.Errorf("expected %s, got %s", JobTypeExtractLabResults, jt)
	}
}

func TestDocumentSubtypeToJobType_LabPDF_ReturnsExtractLabResults(t *testing.T) {
	jt := DocumentSubtypeToJobType(DocumentSubtypeLabPDF)
	if jt != JobTypeExtractLabResults {
		t.Errorf("expected %s, got %s", JobTypeExtractLabResults, jt)
	}
}

func TestDocumentSubtypeToJobType_WearableCSV_ReturnsExtractWearable(t *testing.T) {
	jt := DocumentSubtypeToJobType(DocumentSubtypeWearableCSV)
	if jt != JobTypeExtractWearable {
		t.Errorf("expected %s, got %s", JobTypeExtractWearable, jt)
	}
}

func TestDocumentSubtypeToJobType_WearableJSON_ReturnsExtractWearable(t *testing.T) {
	jt := DocumentSubtypeToJobType(DocumentSubtypeWearableJSON)
	if jt != JobTypeExtractWearable {
		t.Errorf("expected %s, got %s", JobTypeExtractWearable, jt)
	}
}

func TestDocumentSubtypeToJobType_NutritionCSV_ReturnsExtractNutrition(t *testing.T) {
	jt := DocumentSubtypeToJobType(DocumentSubtypeNutritionCSV)
	if jt != JobTypeExtractNutrition {
		t.Errorf("expected %s, got %s", JobTypeExtractNutrition, jt)
	}
}

func TestDocumentSubtypeToJobType_Audio_ReturnsTranscribeAudio(t *testing.T) {
	jt := DocumentSubtypeToJobType(DocumentSubtypeAudio)
	if jt != JobTypeTranscribeAudio {
		t.Errorf("expected %s, got %s", JobTypeTranscribeAudio, jt)
	}
}

func TestDocumentSubtypeToJobType_Other_ReturnsClassifyDocument(t *testing.T) {
	jt := DocumentSubtypeToJobType(DocumentSubtypeOther)
	if jt != JobTypeClassifyDocument {
		t.Errorf("expected %s, got %s", JobTypeClassifyDocument, jt)
	}
}
