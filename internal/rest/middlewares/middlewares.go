package middlewares

import (
	"context"
	"net/http"

	"github.com/netbill/auth-svc/internal/rest/contexter"
	"github.com/netbill/logium"
	"github.com/netbill/restkit/grants"
	"github.com/netbill/restkit/problems"
)

type responser interface {
	Render(w http.ResponseWriter, status int, res ...interface{})
	RenderErr(w http.ResponseWriter, errs ...error)
}

type Provider struct {
	log             *logium.Logger
	accountAccessSK string
	uploadFilesSK   string

	responser responser
}

type Config struct {
	AccountAccessSK string
	UploadFilesSK   string
}

func New(
	log *logium.Logger,
	cfg Config,
	responser responser,
) Provider {
	return Provider{
		accountAccessSK: cfg.AccountAccessSK,
		uploadFilesSK:   cfg.UploadFilesSK,
		log:             log,
		responser:       responser,
	}
}

func (p Provider) AccountAuth(
	allowedRoles ...string,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res, err := grants.AccountAuthToken(
				r,
				p.accountAccessSK,
				"",
				allowedRoles...,
			)
			if err != nil {
				p.log.WithError(err).Errorf("account authentication failed")
				p.responser.RenderErr(w, problems.Unauthorized("account authentication failed"))

				return
			}

			next.ServeHTTP(w, r.WithContext(
				context.WithValue(r.Context(), contexter.AccountDataCtxKey, res)),
			)
		})
	}
}
