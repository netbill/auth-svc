package middlewares

import (
	"github.com/netbill/restkit/tokens"
)

type tokenManager interface {
	ParseUserAuthAccess(tokenStr string) (tokens.AccountAuthClaims, error)
}

type Provider struct {
	tokenManager tokenManager
}

func New(
	tokenManager tokenManager,
) *Provider {
	return &Provider{
		tokenManager: tokenManager,
	}
}
