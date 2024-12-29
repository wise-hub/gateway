package util

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateUniqueToken() (string, error) {
	randomBytes := make([]byte, 24)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}
