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

func TestInMemoryFileStorage_DownloadFile_ExistingObject_ReturnsContent(t *testing.T) {
	storage := NewInMemoryFileStorage()
	content := []byte("test file content")
	storage.PutObjectWithContent("test/key.pdf", content)

	data, err := storage.DownloadFile(context.Background(), "test/key.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "test file content" {
		t.Errorf("expected content %q, got %q", "test file content", string(data))
	}
}

func TestInMemoryFileStorage_DownloadFile_NonExistent_ReturnsError(t *testing.T) {
	storage := NewInMemoryFileStorage()

	_, err := storage.DownloadFile(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent object")
	}
}

func TestInMemoryFileStorage_PutObjectWithContent_AlsoSetsObjectExists(t *testing.T) {
	storage := NewInMemoryFileStorage()
	storage.PutObjectWithContent("test/key.pdf", []byte("content"))

	exists, err := storage.ObjectExists(context.Background(), "test/key.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected object to exist after PutObjectWithContent")
	}
}
