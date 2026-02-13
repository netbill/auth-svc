package middlewares

import (
	"net/http"

	"github.com/netbill/restkit/tokens"
)

type responser interface {
	Render(w http.ResponseWriter, status int, res ...interface{})
	RenderErr(w http.ResponseWriter, errs ...error)
}

type tokenManager interface {
	ParseAccountAuthAccessClaims(tokenStr string) (tokens.AccountAuthClaims, error)
}

type Provider struct {
	tokenManager tokenManager
	responser    responser
}

func New(
	tokenManager tokenManager,
	responser responser,
) *Provider {
	return &Provider{
		tokenManager: tokenManager,
		responser:    responser,
	}
}
