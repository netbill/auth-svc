package tokenmanger

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

const (
	AuthActor = "auth-svc"
)

type Manager struct {
	accessSK  string
	refreshSK string
	refreshHK string

	accessTTL  time.Duration
	refreshTTL time.Duration

	iss string
}

type NewParams struct {
	AccessSK  string
	RefreshSK string
	RefreshHK string

	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewManager(params NewParams) *Manager {
	return &Manager{
		accessSK:   params.AccessSK,
		refreshSK:  params.RefreshSK,
		refreshHK:  params.RefreshHK,
		accessTTL:  params.AccessTTL,
		refreshTTL: params.RefreshTTL,
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
