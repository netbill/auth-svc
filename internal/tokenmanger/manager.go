package tokenmanger

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

type Manager struct {
	AccessSK  string
	RefreshSK string
	RefreshHK string

	AccessTTL  time.Duration
	RefreshTTL time.Duration

	Issuer string
}

func generateOpaque(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hmacB64(msg string, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("empty secret")
	}
	m := hmac.New(sha256.New, []byte(secret))
	_, _ = m.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil)), nil
}
