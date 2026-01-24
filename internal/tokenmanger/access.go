package tokenmanger

import (
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/tokens"
)

func (s Service) GenerateAccess(account models.Account, sessionID uuid.UUID) (string, error) {
	return tokens.GenerateAccountJWT(tokens.GenerateAccountJwtRequest{
		Issuer:    s.iss,
		Audience:  []string{s.iss},
		AccountID: account.ID,
		SessionID: sessionID,
		Role:      account.Role,
		Ttl:       s.accessTTL,
	}, s.accessSK)
}

func (s Service) ParseAccessClaims(tokenStr string) (tokens.AccountJwtData, error) {
	return tokens.ParseAccountJWT(tokenStr, s.accessSK)
}
