package tokenmanger

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

type Service struct {
	accessSK  string
	refreshSK string
	refreshHK string

	accessTTL  time.Duration
	refreshTTL time.Duration

	iss string
}

type Config struct {
	AccessSK  string
	RefreshSK string
	RefreshHK string

	AccessTTL  time.Duration
	RefreshTTL time.Duration

	Iss string
}

func NewManager(cfg Config) Service {
	return Service{
		accessSK:   cfg.AccessSK,
		refreshSK:  cfg.RefreshSK,
		refreshHK:  cfg.RefreshHK,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		iss:        cfg.Iss,
	}
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
