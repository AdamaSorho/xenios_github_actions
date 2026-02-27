package repository

import (
	"context"
	"testing"
)

func TestInMemoryFileDownloader_DownloadFile_ReturnsContent(t *testing.T) {
	d := NewInMemoryFileDownloader()
	d.PutFile("key-1", []byte("hello world"))

	data, err := d.DownloadFile(context.Background(), "key-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}
}

func TestInMemoryFileDownloader_DownloadFile_NotFound_ReturnsError(t *testing.T) {
	d := NewInMemoryFileDownloader()

	_, err := d.DownloadFile(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestInMemoryFileDownloader_DownloadFile_ReturnsCopy(t *testing.T) {
	d := NewInMemoryFileDownloader()
	d.PutFile("key-1", []byte("original"))

	data, _ := d.DownloadFile(context.Background(), "key-1")
	data[0] = 'X' // modify returned copy

	data2, _ := d.DownloadFile(context.Background(), "key-1")
	if string(data2) != "original" {
		t.Errorf("expected original data to be unmodified, got %q", string(data2))
	}
}
