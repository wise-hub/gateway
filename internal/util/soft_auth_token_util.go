package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	secretCache = sync.Map{} 
)

func deriveSecretFromUserAgent(userAgent string) string {
	if secret, ok := secretCache.Load(userAgent); ok {
		return secret.(string)
	}

	indices := []int{1, 3, 7, 9, 11, 13, 15, 17}
	extracted := make([]byte, 0, len(indices))
	for _, idx := range indices {
		if idx-1 < len(userAgent) {
			extracted = append(extracted, userAgent[idx-1])
		}
	}

	h := sha256.Sum256(extracted)
	secret := base64.RawURLEncoding.EncodeToString(h[:])
	secretCache.Store(userAgent, secret)
	return secret
}

func GenerateSoftAuthToken(userAgent string) string {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	secret := deriveSecretFromUserAgent(userAgent)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(timestamp))
	h.Write([]byte(userAgent)) 
	hmacSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return timestamp + "." + hmacSignature
}

func ValidateSoftAuthToken(token, userAgent string) bool {

	if token == "" || userAgent == "" {
		return false
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || time.Now().UnixMilli()-timestamp > 600*60*1000 {
		return false
	}

	h := hmac.New(sha256.New, []byte(deriveSecretFromUserAgent(userAgent)))
	h.Write([]byte(parts[0]))
	h.Write([]byte(userAgent)) 
	expectedHmac := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return parts[1] == expectedHmac
}
