package entities

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// ArtifactStatus represents the upload lifecycle of an artifact.
type ArtifactStatus string

const (
	ArtifactStatusPending  ArtifactStatus = "pending"
	ArtifactStatusUploaded ArtifactStatus = "uploaded"
	ArtifactStatusFailed    ArtifactStatus = "failed"
	ArtifactStatusProcessed ArtifactStatus = "processed"
)

// ArtifactType categorizes the kind of file being stored.
type ArtifactType string

const (
	ArtifactTypeDocument ArtifactType = "document"
	ArtifactTypeAudio    ArtifactType = "audio"
	ArtifactTypeImage    ArtifactType = "image"
	ArtifactTypeVideo    ArtifactType = "video"
)

// MaxDocumentSizeBytes is the maximum file size for documents (10MB).
const MaxDocumentSizeBytes int64 = 10 * 1024 * 1024

// MaxAudioVideoSizeBytes is the maximum file size for audio/video (100MB).
const MaxAudioVideoSizeBytes int64 = 100 * 1024 * 1024

// MaxImageSizeBytes is the maximum file size for images (10MB).
const MaxImageSizeBytes int64 = 10 * 1024 * 1024

// PresignedURLExpiryMinutes is the default expiry for presigned URLs.
const PresignedURLExpiryMinutes = 15

// AllowedFileExtensions maps file extensions to their artifact types.
var AllowedFileExtensions = map[string]ArtifactType{
	".pdf":  ArtifactTypeDocument,
	".csv":  ArtifactTypeDocument,
	".json": ArtifactTypeDocument,
	".xml":  ArtifactTypeDocument,
	".aac":  ArtifactTypeAudio,
	".wav":  ArtifactTypeAudio,
	".mp3":  ArtifactTypeAudio,
	".mp4":  ArtifactTypeVideo,
	".jpg":  ArtifactTypeImage,
	".jpeg": ArtifactTypeImage,
	".png":  ArtifactTypeImage,
}

// AllowedContentTypes maps MIME types to their artifact types.
var AllowedContentTypes = map[string]ArtifactType{
	"application/pdf":  ArtifactTypeDocument,
	"text/csv":         ArtifactTypeDocument,
	"application/json": ArtifactTypeDocument,
	"application/xml":  ArtifactTypeDocument,
	"text/xml":         ArtifactTypeDocument,
	"audio/aac":        ArtifactTypeAudio,
	"audio/wav":        ArtifactTypeAudio,
	"audio/mpeg":       ArtifactTypeAudio,
	"video/mp4":        ArtifactTypeVideo,
	"image/jpeg":       ArtifactTypeImage,
	"image/png":        ArtifactTypeImage,
}

// Artifact represents a file stored in the object storage system.
type Artifact struct {
	ID          string         `json:"id"`
	ClientID    string         `json:"client_id"`
	CoachID     string         `json:"coach_id"`
	FileName    string         `json:"file_name"`
	FileType    string         `json:"file_type"`
	FileSize    int64          `json:"file_size"`
	StorageKey  string         `json:"storage_key"`
	Type        ArtifactType   `json:"type"`
	Status      ArtifactStatus `json:"status"`
	ContentType string         `json:"content_type"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ValidateFileExtension checks if the file extension is allowed and returns the artifact type.
func ValidateFileExtension(fileName string) (ArtifactType, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		return "", NewValidationError("file has no extension")
	}
	artType, ok := AllowedFileExtensions[ext]
	if !ok {
		return "", NewValidationError("file type not allowed: %s", ext)
	}
	return artType, nil
}

// ValidateContentType checks if the content type is allowed and returns the artifact type.
func ValidateContentType(contentType string) (ArtifactType, error) {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	artType, ok := AllowedContentTypes[ct]
	if !ok {
		return "", NewValidationError("content type not allowed: %s", ct)
	}
	return artType, nil
}

// ValidateFileSize checks if the file size is within the allowed limit for the artifact type.
func ValidateFileSize(size int64, artType ArtifactType) error {
	if size <= 0 {
		return NewValidationError("file size must be positive")
	}
	var maxSize int64
	switch artType {
	case ArtifactTypeDocument:
		maxSize = MaxDocumentSizeBytes
	case ArtifactTypeAudio, ArtifactTypeVideo:
		maxSize = MaxAudioVideoSizeBytes
	case ArtifactTypeImage:
		maxSize = MaxImageSizeBytes
	default:
		return NewValidationError("unknown artifact type: %s", artType)
	}
	if size > maxSize {
		return NewValidationError("file size %d exceeds maximum %d bytes for %s", size, maxSize, artType)
	}
	return nil
}

// BuildStorageKey constructs the S3 object key for a file.
// Format: {client_id}/{artifact_type}/{artifact_id}{ext}
func BuildStorageKey(clientID string, artType ArtifactType, artifactID string, fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	return fmt.Sprintf("%s/%s/%s%s", clientID, artType, artifactID, ext)
}
