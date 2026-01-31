package tokenmanger

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/tokens"
)

func (s Service) GenerateAccess(account models.Account, sessionID uuid.UUID) (string, error) {
	tkn, err := tokens.AccountClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account.ID.String(),
			Issuer:    AuthActor,
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(s.accessTTL)),
		},
		Role:      account.Role,
		SessionID: sessionID,
	}.GenerateJWT(s.accessSK)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token, cause: %w", err)
	}

	return tkn, nil
}

func (s Service) ParseAccessClaims(tokenStr string) (tokens.AccountClaims, error) {
	data, err := tokens.ParseAccountJWT(tokenStr, s.accessSK)
	if err != nil {
		return tokens.AccountClaims{}, fmt.Errorf("failed to parse access token, cause: %w", err)
	}

	return data, nil
}
