package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON_Success(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	data := map[string]string{"message": "success"}

	// Act
	err := respondJSON(rec, http.StatusOK, data)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["message"] != "success" {
		t.Errorf("expected message 'success', got %q", response["message"])
	}
}

func TestRespondJSON_SetsContentTypeHeader(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	data := map[string]string{"test": "value"}

	// Act
	err := respondJSON(rec, http.StatusOK, data)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestRespondJSON_SetsStatusCode(t *testing.T) {
	// Arrange
	testCases := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, status := range testCases {
		t.Run(http.StatusText(status), func(t *testing.T) {
			rec := httptest.NewRecorder()
			data := map[string]string{"status": http.StatusText(status)}

			// Act
			err := respondJSON(rec, status, data)

			// Assert
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if rec.Code != status {
				t.Errorf("expected status %d, got %d", status, rec.Code)
			}
		})
	}
}

func TestRespondJSON_HandlesNilData(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()

	// Act
	err := respondJSON(rec, http.StatusOK, nil)

	// Assert
	if err != nil {
		t.Errorf("expected no error for nil data, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// nil encodes to "null" in JSON
	if rec.Body.String() != "null\n" {
		t.Errorf("expected 'null\\n', got %q", rec.Body.String())
	}
}

func TestRespondJSON_HandlesEmptyStruct(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	data := struct{}{}

	// Act
	err := respondJSON(rec, http.StatusOK, data)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// empty struct encodes to "{}"
	if rec.Body.String() != "{}\n" {
		t.Errorf("expected '{}\\n', got %q", rec.Body.String())
	}
}

func TestRespondJSON_HandlesComplexData(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	type ComplexData struct {
		ID      int                    `json:"id"`
		Name    string                 `json:"name"`
		Active  bool                   `json:"active"`
		Tags    []string               `json:"tags"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	data := ComplexData{
		ID:     123,
		Name:   "test",
		Active: true,
		Tags:   []string{"tag1", "tag2"},
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	// Act
	err := respondJSON(rec, http.StatusOK, data)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var response ComplexData
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != data.ID {
		t.Errorf("expected ID %d, got %d", data.ID, response.ID)
	}
}

// Error Handling Tests

type failingWriter struct {
	headerWritten bool
	writeErr      error
}

func (f *failingWriter) Header() http.Header {
	return http.Header{}
}

func (f *failingWriter) Write([]byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return 0, errors.New("write failed")
}

func (f *failingWriter) WriteHeader(statusCode int) {
	f.headerWritten = true
}

func TestRespondJSON_HandlesWriteError(t *testing.T) {
	// Arrange
	fw := &failingWriter{}
	data := map[string]string{"message": "test"}

	// Act
	err := respondJSON(fw, http.StatusOK, data)

	// Assert - should return error when write fails
	if err == nil {
		t.Error("expected error when ResponseWriter.Write fails")
	}

	if !fw.headerWritten {
		t.Error("expected WriteHeader to be called before Write")
	}
}

type unencodeableData struct {
	Channel chan int // channels cannot be encoded to JSON
}

func TestRespondJSON_HandlesEncodingError(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	data := unencodeableData{Channel: make(chan int)}

	// Act
	err := respondJSON(rec, http.StatusOK, data)

	// Assert - should return error when encoding fails
	if err == nil {
		t.Error("expected error when JSON encoding fails")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d even on encoding error, got %d", http.StatusOK, rec.Code)
	}
}

func TestRespondJSON_HandlesCircularReference(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	type Node struct {
		Value int
		Next  *Node
	}

	// Create circular reference
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node1.Next = node2
	node2.Next = node1

	// Act
	err := respondJSON(rec, http.StatusOK, node1)

	// Assert - JSON encoder will fail on circular references
	if err == nil {
		t.Error("expected error for circular reference")
	}
}

func TestRespondError_Success(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	message := "test error message"

	// Act
	respondError(rec, http.StatusBadRequest, message)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}

	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != message {
		t.Errorf("expected error %q, got %q", message, response.Error)
	}
}

func TestRespondError_EmptyMessage(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()

	// Act
	respondError(rec, http.StatusInternalServerError, "")

	// Assert
	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "" {
		t.Errorf("expected empty error message, got %q", response.Error)
	}
}

func TestRespondError_LongMessage(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	longMessage := string(make([]byte, 10000))
	for i := range longMessage {
		longMessage = longMessage[:i] + "a" + longMessage[i+1:]
	}

	// Act
	respondError(rec, http.StatusBadRequest, longMessage)

	// Assert
	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != longMessage {
		t.Errorf("expected long error message to be preserved")
	}
}

func TestRespondError_SpecialCharacters(t *testing.T) {
	// Arrange
	testMessages := []string{
		"error with \"quotes\"",
		"error with 'single quotes'",
		"error with\nnewlines",
		"error with\ttabs",
		"error with unicode 🚀",
		"error with backslash \\",
	}

	for _, msg := range testMessages {
		t.Run(msg, func(t *testing.T) {
			rec := httptest.NewRecorder()

			// Act
			respondError(rec, http.StatusBadRequest, msg)

			// Assert
			var response ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Error != msg {
				t.Errorf("expected error %q, got %q", msg, response.Error)
			}
		})
	}
}

func TestErrorResponse_JSONStructure(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	message := "test error"

	// Act
	respondError(rec, http.StatusBadRequest, message)

	// Assert - verify JSON structure
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&rawResponse); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, exists := rawResponse["error"]; !exists {
		t.Error("expected 'error' field in response")
	}

	if len(rawResponse) != 1 {
		t.Errorf("expected exactly 1 field, got %d", len(rawResponse))
	}
}

func TestRespondErrorWithCode_FullResponse(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	details := map[string]interface{}{
		"field": "email",
		"reason": "invalid format",
	}

	// Act
	respondErrorWithCode(rec, http.StatusBadRequest, "validation failed", "VALIDATION_ERROR", details)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "validation failed" {
		t.Errorf("expected error 'validation failed', got %q", response.Error)
	}
	if response.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got %q", response.Code)
	}
	if response.Details == nil {
		t.Fatal("expected details to be non-nil")
	}
	if response.Details["field"] != "email" {
		t.Errorf("expected details.field 'email', got %v", response.Details["field"])
	}
}

func TestRespondErrorWithCode_NilDetails(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()

	// Act
	respondErrorWithCode(rec, http.StatusNotFound, "not found", "NOT_FOUND", nil)

	// Assert
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&rawResponse); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// With nil details and omitempty, the "details" field should not be present
	if _, exists := rawResponse["details"]; exists {
		t.Error("expected 'details' field to be omitted when nil")
	}
	if rawResponse["error"] != "not found" {
		t.Errorf("expected error 'not found', got %v", rawResponse["error"])
	}
	if rawResponse["code"] != "NOT_FOUND" {
		t.Errorf("expected code 'NOT_FOUND', got %v", rawResponse["code"])
	}
}
