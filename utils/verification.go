package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateVerificationCode menghasilkan kode verifikasi 6 digit
func GenerateVerificationCode() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:6], nil
}
