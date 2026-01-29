package tokenmanger

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/tokens"
)

func (s Service) GenerateRefresh(account models.Account, sessionID uuid.UUID) (string, error) {
	tkn, err := tokens.GenerateAccountJWT(tokens.GenerateAccountJwtRequest{
		Issuer: s.iss,
		//TODO add audience when needed for refresh tokens this may be useful
		//Audience:  []string{s.iss},
		AccountID: account.ID,
		SessionID: sessionID,
		Role:      account.Role,
		Ttl:       s.refreshTTL,
	}, s.refreshSK)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token, cause: %w", err)
	}

	return tkn, nil
}

func (s Service) ParseRefreshClaims(tokenStr string) (tokens.AccountJwtData, error) {
	data, err := tokens.ParseAccountJWT(tokenStr, s.refreshSK)
	if err != nil {
		return tokens.AccountJwtData{}, fmt.Errorf("failed to parse refresh token, cause: %w", err)
	}

	return data, nil
}

func (s Service) HashRefresh(rawRefresh string) (string, error) {
	hash, err := hmacB64("refresh."+rawRefresh, s.refreshHK)
	if err != nil {
		return "", fmt.Errorf("failed to hash refresh token, cause: %w", err)
	}

	return hash, nil
}
