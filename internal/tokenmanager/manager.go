package tokenmanager

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

type Config struct {
	Issuer                  string
	AccountAccessSecretKey  string
	AccountAccessTTL        time.Duration
	AccountRefreshSecretKey string
	AccountRefreshHashKey   string
	AccountRefreshTTL       time.Duration
}

type Manager struct {
	issuer string

	accessSK  string
	refreshSK string
	refreshHK string

	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(config Config) *Manager {
	return &Manager{
		issuer:     config.Issuer,
		accessSK:   config.AccountAccessSecretKey,
		refreshSK:  config.AccountRefreshSecretKey,
		refreshHK:  config.AccountRefreshHashKey,
		accessTTL:  config.AccountAccessTTL,
		refreshTTL: config.AccountRefreshTTL,
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
