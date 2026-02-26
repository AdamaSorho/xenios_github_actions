package repository

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestInMemoryFileStorage_GenerateUploadURL_ReturnsURL(t *testing.T) {
	storage := NewInMemoryFileStorage()

	result, err := storage.GenerateUploadURL(context.Background(), "client-1/document/test.pdf", "application/pdf", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL == "" {
		t.Fatal("expected non-empty URL")
	}
	if !strings.Contains(result.URL, "client-1/document/test.pdf") {
		t.Errorf("expected URL to contain key, got %s", result.URL)
	}
}

func TestInMemoryFileStorage_GenerateUploadURL_SetsExpiry(t *testing.T) {
	storage := NewInMemoryFileStorage()

	result, err := storage.GenerateUploadURL(context.Background(), "test/key", "application/pdf", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExpiresAt.Before(time.Now()) {
		t.Error("expected expiry to be in the future")
	}
}

func TestInMemoryFileStorage_GenerateDownloadURL_ExistingObject_ReturnsURL(t *testing.T) {
	storage := NewInMemoryFileStorage()
	storage.PutObject("client-1/document/test.pdf")

	result, err := storage.GenerateDownloadURL(context.Background(), "client-1/document/test.pdf", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL == "" {
		t.Fatal("expected non-empty URL")
	}
}

func TestInMemoryFileStorage_GenerateDownloadURL_NonExistent_ReturnsError(t *testing.T) {
	storage := NewInMemoryFileStorage()

	_, err := storage.GenerateDownloadURL(context.Background(), "nonexistent", 15*time.Minute)
	if err == nil {
		t.Fatal("expected error for nonexistent object")
	}
}

func TestInMemoryFileStorage_ObjectExists_Exists_ReturnsTrue(t *testing.T) {
	storage := NewInMemoryFileStorage()
	storage.PutObject("test/key")

	exists, err := storage.ObjectExists(context.Background(), "test/key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected object to exist")
	}
}

func TestInMemoryFileStorage_ObjectExists_NotExists_ReturnsFalse(t *testing.T) {
	storage := NewInMemoryFileStorage()

	exists, err := storage.ObjectExists(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected object to not exist")
	}
}

func TestInMemoryFileStorage_DownloadFile_ExistingObject_ReturnsData(t *testing.T) {
	storage := NewInMemoryFileStorage()
	storage.PutObjectWithData("test/file.pdf", []byte("pdf-content"))

	data, err := storage.DownloadFile(context.Background(), "test/file.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "pdf-content" {
		t.Errorf("expected 'pdf-content', got %q", string(data))
	}
}

func TestInMemoryFileStorage_DownloadFile_NonExistent_ReturnsError(t *testing.T) {
	storage := NewInMemoryFileStorage()

	_, err := storage.DownloadFile(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent object")
	}
}

func TestInMemoryFileStorage_PutObjectWithData_SetsObjectAndData(t *testing.T) {
	storage := NewInMemoryFileStorage()
	storage.PutObjectWithData("key", []byte("data"))

	exists, _ := storage.ObjectExists(context.Background(), "key")
	if !exists {
		t.Error("expected object to exist")
	}

	data, err := storage.DownloadFile(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "data" {
		t.Errorf("expected 'data', got %q", string(data))
	}
}
