package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
)

const operationDeleteMySessions = "delete_my_sessions"

func (c *Controller) DeleteMySessions(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMySessions)

	err := c.core.DeleteMySessions(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound) || errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Infof("account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("account not found by credentials"))
	case err != nil:
		log.WithError(err).Error("failed to delete my sessions")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		log.Info("sessions deleted")
		c.responser.Status(w, http.StatusNoContent)
	}
}
