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
