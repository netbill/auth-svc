package middlewares

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/headers"
	"github.com/netbill/restkit/problems"
)

func (p *Provider) AccountAuth(allowedRoles ...string) func(next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := headers.GetAuthorizationToken(r)
			if err != nil {
				scope.Log(r).WithError(err).Debug("account authentication failed")
				p.responser.RenderErr(w, problems.Unauthorized("account authentication failed"))

				return
			}

			claims, err := p.tokenManager.ParseAccountAuthAccess(token)
			if err != nil {
				scope.Log(r).WithError(err).Info("account authentication failed")
				p.responser.RenderErr(w, problems.Unauthorized("account authentication failed"))

				return
			}

			if len(allowed) > 0 {
				if _, ok := allowed[claims.Role]; !ok {
					scope.Log(r).Debug("account authentication rejected by role")
					p.responser.RenderErr(w, problems.Unauthorized("invalid authentication role"))

					return
				}
			}

			next.ServeHTTP(w, r.WithContext(scope.CtxAccountAuth(r.Context(), claims)))
		})
	}
}
