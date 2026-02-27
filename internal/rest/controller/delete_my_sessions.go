package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationDeleteMySessions = "delete_my_sessions"

func (c *Controller) DeleteMySessions(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMySessions)

	err := c.core.DeleteMySessions(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound) || errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Info("invalid credentials")
		render.ResponseError(w, problems.Unauthorized("invalid credentials"))
	case err != nil:
		log.WithError(err).Error("failed to delete my sessions")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("sessions deleted")
		render.Response(w, http.StatusNoContent, nil)
	}
}
