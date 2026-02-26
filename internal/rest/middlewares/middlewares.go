package middlewares

import (
	"net/http"

	"github.com/netbill/restkit/tokens"
)

type responser interface {
	Status(w http.ResponseWriter, status int)
	Render(w http.ResponseWriter, status int, res interface{})
	RenderErr(w http.ResponseWriter, errs ...error)
}

type tokenManager interface {
	ParseAccountAuthAccess(tokenStr string) (tokens.AccountAuthClaims, error)
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
