package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// generateID produces a random hex-encoded 128-bit identifier.
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
