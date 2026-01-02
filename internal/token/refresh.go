package token

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/domain/models"
	"github.com/netbill/restkit/token"
)

func (s Service) GenerateRefresh(account models.Account, sessionID uuid.UUID) (string, error) {
	return token.GenerateAccountJWT(token.GenerateAccountJwtRequest{
		Issuer:    s.iss,
		Audience:  []string{s.iss},
		AccountID: account.ID,
		SessionID: sessionID,
		Username:  account.Username,
		Role:      account.Role,
		Ttl:       s.refreshTTL,
	}, s.refreshSK)
}

func (s Service) EncryptRefresh(token string) (string, error) {
	return encryptAESGCM(token, []byte(s.refreshSK))
}

func (s Service) DecryptRefresh(encryptedToken string) (string, error) {
	raw, err := decryptAESGCM(encryptedToken, []byte(s.refreshSK))
	if err != nil {
		return "", fmt.Errorf("decrypt refresh: %w", err)
	}

	return raw, nil
}

func (s Service) ParseRefreshClaims(tokenStr string) (token.AccountClaims, error) {
	return token.VerifyAccountJWT(tokenStr, s.refreshSK)
}
