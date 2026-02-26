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

func TestInMemoryFileStorage_PutObjectWithContent_StoresContent(t *testing.T) {
	storage := NewInMemoryFileStorage()
	content := []byte("hello world")
	storage.PutObjectWithContent("test/key", content)

	exists, err := storage.ObjectExists(context.Background(), "test/key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected object to exist after PutObjectWithContent")
	}
}

func TestInMemoryFileStorage_DownloadObject_ExistingKey_ReturnsContent(t *testing.T) {
	storage := NewInMemoryFileStorage()
	content := []byte("csv,data,here")
	storage.PutObjectWithContent("test/nutrition.csv", content)

	data, err := storage.DownloadObject(context.Background(), "test/nutrition.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("expected %q, got %q", string(content), string(data))
	}
}

func TestInMemoryFileStorage_DownloadObject_NonExistent_ReturnsError(t *testing.T) {
	storage := NewInMemoryFileStorage()

	_, err := storage.DownloadObject(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent object")
	}
}

func TestInMemoryFileStorage_DownloadObject_ReturnsCopy(t *testing.T) {
	storage := NewInMemoryFileStorage()
	content := []byte("original")
	storage.PutObjectWithContent("test/key", content)

	data, err := storage.DownloadObject(context.Background(), "test/key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Modify the returned data
	data[0] = 'X'

	// Verify original is unchanged
	data2, _ := storage.DownloadObject(context.Background(), "test/key")
	if string(data2) != "original" {
		t.Errorf("expected original content to be unchanged, got %q", string(data2))
	}
}
