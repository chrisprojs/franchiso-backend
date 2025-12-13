package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateVerificationCode generates a 6-digit verification code
func GenerateVerificationCode() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:6], nil
}
