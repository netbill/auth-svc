package controller

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
)

const operationLogout = "logout"

func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationLogout)

	err := c.core.Logout(r.Context(), scope.AccountActor(r))
	switch {
	case err != nil:
		log.WithError(err).Error("logout failed")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		c.responser.Status(w, http.StatusNoContent)
	}
}
