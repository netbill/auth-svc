package tokenmanger

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

const Issuer = "auth-svc"

type Config struct {
	AccountAccess struct {
		SecretKey string        `mapstructure:"secret_key"`
		TTL       time.Duration `mapstructure:"ttl"`
	} `mapstructure:"account_access"`
	AccountRefresh struct {
		SecretKey string        `mapstructure:"secret_key"`
		HashKey   string        `mapstructure:"hash_key"`
		TTL       time.Duration `mapstructure:"ttl"`
	} `mapstructure:"account_refresh"`
}

type Manager struct {
	Issuer string

	accessSK  string
	refreshSK string
	refreshHK string

	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(issuer string, config Config) *Manager {
	return &Manager{
		Issuer:     Issuer,
		accessSK:   config.AccountAccess.SecretKey,
		refreshSK:  config.AccountRefresh.SecretKey,
		refreshHK:  config.AccountRefresh.HashKey,
		accessTTL:  config.AccountAccess.TTL,
		refreshTTL: config.AccountRefresh.TTL,
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
