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

	err := c.core.DeleteOwnSessions(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorInitiatorNotFound):
		log.Info("initiator account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
	case errors.Is(err, errx.ErrorInitiatorInvalidSession):
		log.Info("initiator session is invalid")
		c.responser.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
	case err != nil:
		log.WithError(err).Error("failed to delete my sessions")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		log.Info("initiator sessions deleted")
		c.responser.Render(w, http.StatusNoContent)
	}
}
