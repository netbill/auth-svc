package token

import (
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/domain/models"
	"github.com/netbill/restkit/token"
)

func (s Service) EncryptAccess(token string) (string, error) {
	return encryptAESGCM(token, []byte(s.accessSK))
}

func (s Service) GenerateAccess(user models.Account, sessionID uuid.UUID) (string, error) {
	return token.GenerateAccountJWT(token.GenerateAccountJwtRequest{
		Issuer:    s.iss,
		AccountID: user.ID,
		//Audience:  []string{"gateway"},
		SessionID: sessionID,
		Role:      user.Role,
		Username:  user.Username,
		Ttl:       s.accessTTL,
	}, s.accessSK)
}
