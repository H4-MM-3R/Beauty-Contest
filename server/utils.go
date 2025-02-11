package server

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateHubHash returns a random 8-character hex string.
func GenerateHubHash() (string, error) {
	bytes := make([]byte, 4) // 4 bytes -> 8 hex characters.
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

