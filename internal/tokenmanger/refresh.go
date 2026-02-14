package tokenmanger

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/tokens"
)

func (m *Manager) GenerateRefresh(account models.Account, sessionID uuid.UUID) (string, error) {
	tkn, err := tokens.AccountAuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account.ID.String(),
			Issuer:    m.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(m.accessTTL)),
		},
		Role:      account.Role,
		SessionID: sessionID,
	}.GenerateJWT(m.refreshSK)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token, cause: %w", err)
	}

	return tkn, nil
}

func (m *Manager) ParseAccountAuthRefreshClaims(tokenStr string) (tokens.AccountAuthClaims, error) {
	data, err := tokens.ParseAccountJWT(tokenStr, m.refreshSK)
	if err != nil {
		return tokens.AccountAuthClaims{}, fmt.Errorf("failed to parse refresh token, cause: %w", err)
	}

	return data, nil
}

func (m *Manager) HashRefresh(rawRefresh string) (string, error) {
	hash, err := hmacB64("refresh."+rawRefresh, m.refreshHK)
	if err != nil {
		return "", fmt.Errorf("failed to hash refresh token, cause: %w", err)
	}

	return hash, nil
}
