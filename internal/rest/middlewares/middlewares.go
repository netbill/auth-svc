package middlewares

import (
	"net/http"

	"github.com/netbill/logium"
	"github.com/netbill/restkit/mdlv"
)

type Service struct {
	accountAccessSK string

	log logium.Logger
}

type Config struct {
	AccountAccessSK string
}

func New(
	log logium.Logger,
	accountAccessSK string,
) Service {
	return Service{
		accountAccessSK: accountAccessSK,
		log:             log,
	}
}

func (s Service) AccountAuth() func(next http.Handler) http.Handler {
	return mdlv.AccountAuth(s.log, accountDataCtxKey, s.accountAccessSK)
}

func (s Service) AccountRoleGrant(
	allowedRoles map[string]bool,
) func(http.Handler) http.Handler {
	return mdlv.AccountRoleGrant(s.log, accountDataCtxKey, allowedRoles)
}
