package tokenmanager

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
)

type Config struct {
	Issuer string

	AccessSecretKey string
	AccessTTL       time.Duration

	RefreshSecretKey string
	RefreshTTL       time.Duration
	RefreshHashKey   string
}

type Manager struct {
	cfg Config
}

func New(cfg Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) GenerateAccess(user models.User, sessionID uuid.UUID) (string, error) {
	claims := tokens.AccountAuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    m.cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.cfg.AccessTTL)),
		},
		Role:      user.Role,
		SessionID: sessionID,
	}
	return claims.GenerateJWT(m.cfg.AccessSecretKey)
}

func (m *Manager) GenerateRefresh(user models.User, sessionID uuid.UUID) (string, error) {
	claims := tokens.AccountAuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    m.cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.cfg.RefreshTTL)),
		},
		Role:      user.Role,
		SessionID: sessionID,
	}
	return claims.GenerateJWT(m.cfg.RefreshSecretKey)
}

func (m *Manager) ParseUserAuthAccess(tokenStr string) (tokens.AccountAuthClaims, error) {
	return tokens.ParseAccountJWT(tokenStr, m.cfg.AccessSecretKey)
}

func (m *Manager) ParseUserAuthRefresh(tokenStr string) (tokens.AccountAuthClaims, error) {
	return tokens.ParseAccountJWT(tokenStr, m.cfg.RefreshSecretKey)
}

func (m *Manager) HashRefresh(token string) (string, error) {
	mac := hmac.New(sha256.New, []byte(m.cfg.RefreshHashKey))
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
