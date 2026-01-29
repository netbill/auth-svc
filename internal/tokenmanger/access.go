package tokenmanger

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/tokens"
)

func (s Service) GenerateAccess(account models.Account, sessionID uuid.UUID) (string, error) {
	tkn, err := tokens.GenerateAccountJWT(tokens.GenerateAccountJwtRequest{
		Issuer: s.iss,
		//TODO add audience when needed
		//Audience:  []string{s.iss},
		AccountID: account.ID,
		SessionID: sessionID,
		Role:      account.Role,
		Ttl:       s.accessTTL,
	}, s.accessSK)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token, cause: %w", err)
	}

	return tkn, nil
}

func (s Service) ParseAccessClaims(tokenStr string) (tokens.AccountJwtData, error) {
	data, err := tokens.ParseAccountJWT(tokenStr, s.accessSK)
	if err != nil {
		return tokens.AccountJwtData{}, fmt.Errorf("failed to parse access token, cause: %w", err)
	}

	return data, nil
}
