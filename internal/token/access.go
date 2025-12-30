package token

import (
	"github.com/google/uuid"
	"github.com/umisto/restkit/token"
	"github.com/umisto/sso-svc/internal/domain/models"
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
