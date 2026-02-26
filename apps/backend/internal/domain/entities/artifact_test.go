package entities

import (
	"testing"
)

func TestValidateFileExtension_PDF_ReturnsDocument(t *testing.T) {
	artType, err := ValidateFileExtension("report.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artType != ArtifactTypeDocument {
		t.Errorf("expected %s, got %s", ArtifactTypeDocument, artType)
	}
}

func TestValidateFileExtension_CSV_ReturnsDocument(t *testing.T) {
	artType, err := ValidateFileExtension("data.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artType != ArtifactTypeDocument {
		t.Errorf("expected %s, got %s", ArtifactTypeDocument, artType)
	}
}

func TestValidateFileExtension_MP3_ReturnsAudio(t *testing.T) {
	artType, err := ValidateFileExtension("recording.mp3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artType != ArtifactTypeAudio {
		t.Errorf("expected %s, got %s", ArtifactTypeAudio, artType)
	}
}

func TestValidateFileExtension_WAV_ReturnsAudio(t *testing.T) {
	artType, err := ValidateFileExtension("session.wav")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artType != ArtifactTypeAudio {
		t.Errorf("expected %s, got %s", ArtifactTypeAudio, artType)
	}
}

func TestValidateFileExtension_MP4_ReturnsVideo(t *testing.T) {
	artType, err := ValidateFileExtension("video.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artType != ArtifactTypeVideo {
		t.Errorf("expected %s, got %s", ArtifactTypeVideo, artType)
	}
}

func TestValidateFileExtension_PNG_ReturnsImage(t *testing.T) {
	artType, err := ValidateFileExtension("photo.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artType != ArtifactTypeImage {
		t.Errorf("expected %s, got %s", ArtifactTypeImage, artType)
	}
}

func TestValidateFileExtension_UpperCase_AcceptsCaseInsensitive(t *testing.T) {
	artType, err := ValidateFileExtension("Report.PDF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artType != ArtifactTypeDocument {
		t.Errorf("expected %s, got %s", ArtifactTypeDocument, artType)
	}
}

func TestValidateFileExtension_NoExtension_ReturnsError(t *testing.T) {
	_, err := ValidateFileExtension("noextension")
	if err == nil {
		t.Fatal("expected error for file with no extension")
	}
}

func TestValidateFileExtension_DisallowedType_ReturnsError(t *testing.T) {
	_, err := ValidateFileExtension("script.exe")
	if err == nil {
		t.Fatal("expected error for disallowed file type")
	}
}

func TestValidateContentType_ApplicationPDF_ReturnsDocument(t *testing.T) {
	artType, err := ValidateContentType("application/pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artType != ArtifactTypeDocument {
		t.Errorf("expected %s, got %s", ArtifactTypeDocument, artType)
	}
}

func TestValidateContentType_AudioMPEG_ReturnsAudio(t *testing.T) {
	artType, err := ValidateContentType("audio/mpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artType != ArtifactTypeAudio {
		t.Errorf("expected %s, got %s", ArtifactTypeAudio, artType)
	}
}

func TestValidateContentType_Unknown_ReturnsError(t *testing.T) {
	_, err := ValidateContentType("application/octet-stream")
	if err == nil {
		t.Fatal("expected error for unknown content type")
	}
}

func TestValidateFileSize_Document_WithinLimit_NoError(t *testing.T) {
	err := ValidateFileSize(5*1024*1024, ArtifactTypeDocument)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFileSize_Document_ExceedsLimit_ReturnsError(t *testing.T) {
	err := ValidateFileSize(11*1024*1024, ArtifactTypeDocument)
	if err == nil {
		t.Fatal("expected error for oversized document")
	}
}

func TestValidateFileSize_Audio_WithinLimit_NoError(t *testing.T) {
	err := ValidateFileSize(50*1024*1024, ArtifactTypeAudio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFileSize_Audio_ExceedsLimit_ReturnsError(t *testing.T) {
	err := ValidateFileSize(101*1024*1024, ArtifactTypeAudio)
	if err == nil {
		t.Fatal("expected error for oversized audio")
	}
}

func TestValidateFileSize_Video_WithinLimit_NoError(t *testing.T) {
	err := ValidateFileSize(90*1024*1024, ArtifactTypeVideo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFileSize_Image_ExceedsLimit_ReturnsError(t *testing.T) {
	err := ValidateFileSize(11*1024*1024, ArtifactTypeImage)
	if err == nil {
		t.Fatal("expected error for oversized image")
	}
}

func TestValidateFileSize_Zero_ReturnsError(t *testing.T) {
	err := ValidateFileSize(0, ArtifactTypeDocument)
	if err == nil {
		t.Fatal("expected error for zero file size")
	}
}

func TestValidateFileSize_Negative_ReturnsError(t *testing.T) {
	err := ValidateFileSize(-1, ArtifactTypeDocument)
	if err == nil {
		t.Fatal("expected error for negative file size")
	}
}

func TestBuildStorageKey_FormatsCorrectly(t *testing.T) {
	key := BuildStorageKey("client-123", ArtifactTypeDocument, "artifact-456", "report.pdf")
	expected := "client-123/document/artifact-456.pdf"
	if key != expected {
		t.Errorf("expected '%s', got '%s'", expected, key)
	}
}

func TestBuildStorageKey_Audio_FormatsCorrectly(t *testing.T) {
	key := BuildStorageKey("client-789", ArtifactTypeAudio, "artifact-001", "session.mp3")
	expected := "client-789/audio/artifact-001.mp3"
	if key != expected {
		t.Errorf("expected '%s', got '%s'", expected, key)
	}
}

func TestBuildStorageKey_UpperCaseExtension_LowercaseInKey(t *testing.T) {
	key := BuildStorageKey("client-123", ArtifactTypeDocument, "artifact-456", "Report.PDF")
	expected := "client-123/document/artifact-456.pdf"
	if key != expected {
		t.Errorf("expected '%s', got '%s'", expected, key)
	}
}

func TestArtifactStatusConstants_CorrectValues(t *testing.T) {
	if ArtifactStatusPending != "pending" {
		t.Errorf("expected 'pending', got '%s'", ArtifactStatusPending)
	}
	if ArtifactStatusUploaded != "uploaded" {
		t.Errorf("expected 'uploaded', got '%s'", ArtifactStatusUploaded)
	}
	if ArtifactStatusFailed != "failed" {
		t.Errorf("expected 'failed', got '%s'", ArtifactStatusFailed)
	}
}

func TestArtifactTypeConstants_CorrectValues(t *testing.T) {
	if ArtifactTypeDocument != "document" {
		t.Errorf("expected 'document', got '%s'", ArtifactTypeDocument)
	}
	if ArtifactTypeAudio != "audio" {
		t.Errorf("expected 'audio', got '%s'", ArtifactTypeAudio)
	}
	if ArtifactTypeImage != "image" {
		t.Errorf("expected 'image', got '%s'", ArtifactTypeImage)
	}
	if ArtifactTypeVideo != "video" {
		t.Errorf("expected 'video', got '%s'", ArtifactTypeVideo)
	}
}

func TestAllowedFileExtensions_AllTypesPresent(t *testing.T) {
	expected := []string{".pdf", ".csv", ".json", ".xml", ".aac", ".wav", ".mp3", ".mp4", ".jpg", ".jpeg", ".png"}
	for _, ext := range expected {
		if _, ok := AllowedFileExtensions[ext]; !ok {
			t.Errorf("expected extension %s to be in AllowedFileExtensions", ext)
		}
	}
}

// --- DocumentSubtype & Classification Tests ---

func TestIsValidDocumentSubtype_ValidTypes_ReturnTrue(t *testing.T) {
	valid := []DocumentSubtype{
		DocumentSubtypeInBodyPDF, DocumentSubtypeLabCSV, DocumentSubtypeLabPDF,
		DocumentSubtypeWearableCSV, DocumentSubtypeWearableJSON,
		DocumentSubtypeNutritionCSV, DocumentSubtypeAudio, DocumentSubtypeOther,
	}
	for _, st := range valid {
		if !IsValidDocumentSubtype(st) {
			t.Errorf("expected %q to be valid", st)
		}
	}
}

func TestIsValidDocumentSubtype_InvalidType_ReturnFalse(t *testing.T) {
	if IsValidDocumentSubtype("bogus_type") {
		t.Error("expected 'bogus_type' to be invalid")
	}
}

func TestClassifyDocument_HintProvided_UsesHint(t *testing.T) {
	result := ClassifyDocument(DocumentSubtypeInBodyPDF, "scan.pdf", "application/pdf")
	if result != DocumentSubtypeInBodyPDF {
		t.Errorf("expected %s, got %s", DocumentSubtypeInBodyPDF, result)
	}
}

func TestClassifyDocument_HintLabCSV_UsesHint(t *testing.T) {
	result := ClassifyDocument(DocumentSubtypeLabCSV, "blood_test.csv", "text/csv")
	if result != DocumentSubtypeLabCSV {
		t.Errorf("expected %s, got %s", DocumentSubtypeLabCSV, result)
	}
}

func TestClassifyDocument_InvalidHint_FallsBackToExtension(t *testing.T) {
	result := ClassifyDocument("invalid_hint", "recording.mp3", "audio/mpeg")
	if result != DocumentSubtypeAudio {
		t.Errorf("expected %s, got %s", DocumentSubtypeAudio, result)
	}
}

func TestClassifyDocument_NoHint_AudioExtension_ReturnsAudio(t *testing.T) {
	tests := []struct {
		fileName    string
		contentType string
	}{
		{"recording.mp3", "audio/mpeg"},
		{"session.wav", "audio/wav"},
		{"note.aac", "audio/aac"},
	}
	for _, tc := range tests {
		result := ClassifyDocument("", tc.fileName, tc.contentType)
		if result != DocumentSubtypeAudio {
			t.Errorf("ClassifyDocument(%q, %q) = %s, want %s", tc.fileName, tc.contentType, result, DocumentSubtypeAudio)
		}
	}
}

func TestClassifyDocument_NoHint_AudioContentType_ReturnsAudio(t *testing.T) {
	result := ClassifyDocument("", "file", "audio/mpeg")
	if result != DocumentSubtypeAudio {
		t.Errorf("expected %s, got %s", DocumentSubtypeAudio, result)
	}
}

func TestClassifyDocument_NoHint_PDF_ReturnsOther(t *testing.T) {
	result := ClassifyDocument("", "report.pdf", "application/pdf")
	if result != DocumentSubtypeOther {
		t.Errorf("expected %s, got %s", DocumentSubtypeOther, result)
	}
}

func TestClassifyDocument_NoHint_CSV_ReturnsOther(t *testing.T) {
	result := ClassifyDocument("", "data.csv", "text/csv")
	if result != DocumentSubtypeOther {
		t.Errorf("expected %s, got %s", DocumentSubtypeOther, result)
	}
}

func TestClassifyDocument_NoHint_JSON_ReturnsOther(t *testing.T) {
	result := ClassifyDocument("", "export.json", "application/json")
	if result != DocumentSubtypeOther {
		t.Errorf("expected %s, got %s", DocumentSubtypeOther, result)
	}
}

func TestClassifyDocument_MismatchedHintAndExtension_HintWins(t *testing.T) {
	// Hint says lab_csv but file is a PDF — hint takes precedence
	result := ClassifyDocument(DocumentSubtypeLabCSV, "report.pdf", "application/pdf")
	if result != DocumentSubtypeLabCSV {
		t.Errorf("expected %s, got %s", DocumentSubtypeLabCSV, result)
	}
}

func TestDocumentSubtypeToJobType_AllMappings(t *testing.T) {
	tests := []struct {
		subtype  DocumentSubtype
		expected JobType
	}{
		{DocumentSubtypeInBodyPDF, JobTypeExtractInBody},
		{DocumentSubtypeLabCSV, JobTypeExtractLabResults},
		{DocumentSubtypeLabPDF, JobTypeExtractLabResults},
		{DocumentSubtypeWearableCSV, JobTypeExtractWearable},
		{DocumentSubtypeWearableJSON, JobTypeExtractWearable},
		{DocumentSubtypeNutritionCSV, JobTypeExtractNutrition},
		{DocumentSubtypeAudio, JobTypeTranscribeAudio},
		{DocumentSubtypeOther, JobTypeClassifyDocument},
	}
	for _, tc := range tests {
		got := DocumentSubtypeToJobType(tc.subtype)
		if got != tc.expected {
			t.Errorf("DocumentSubtypeToJobType(%s) = %s, want %s", tc.subtype, got, tc.expected)
		}
	}
}

func TestDocumentSubtypeConstants_CorrectValues(t *testing.T) {
	if DocumentSubtypeInBodyPDF != "inbody_pdf" {
		t.Errorf("expected 'inbody_pdf', got '%s'", DocumentSubtypeInBodyPDF)
	}
	if DocumentSubtypeLabCSV != "lab_csv" {
		t.Errorf("expected 'lab_csv', got '%s'", DocumentSubtypeLabCSV)
	}
	if DocumentSubtypeOther != "other" {
		t.Errorf("expected 'other', got '%s'", DocumentSubtypeOther)
	}
}
