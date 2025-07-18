package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// GenerateToken creates a signed token for the user using their secret key.
// The token format: base64(username:timestamp:signature)
func GenerateToken(username, secretKey string) (string, error) {
	timestamp := time.Now().UTC().Unix()
	payload := fmt.Sprintf("%s:%d", username, timestamp)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	token := fmt.Sprintf("%s:%s", payload, signature)
	return base64.StdEncoding.EncodeToString([]byte(token)), nil
}

// ValidateToken checks the token's signature and returns the username if valid.
func ValidateToken(token, secretKey string, maxAge time.Duration) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("invalid token encoding")
	}
	parts := string(decoded)
	var username string
	var timestamp int64
	var signature string
	n, err := fmt.Sscanf(parts, "%[^:]:%d:%s", &username, &timestamp, &signature)
	if err != nil || n != 3 {
		return "", fmt.Errorf("invalid token format")
	}
	payload := fmt.Sprintf("%s:%d", username, timestamp)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return "", fmt.Errorf("invalid token signature")
	}
	// Check token age
	issuedAt := time.Unix(timestamp, 0)
	if time.Since(issuedAt) > maxAge {
		return "", fmt.Errorf("token expired")
	}
	return username, nil
}

// DecodeTokenUsername extracts the username from the token without validating the signature.
func DecodeTokenUsername(token string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("invalid token encoding")
	}
	parts := string(decoded)
	var username string
	var timestamp int64
	var signature string
	n, err := fmt.Sscanf(parts, "%[^:]:%d:%s", &username, &timestamp, &signature)
	if err != nil || n != 3 {
		return "", fmt.Errorf("invalid token format")
	}
	return username, nil
}
