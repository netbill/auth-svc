package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
)

const operationDeleteMyAccount = "delete_my_account"

func (c *Controller) DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMyAccount)

	err := c.core.DeleteOwnAccount(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorInitiatorNotFound):
		log.Info("initiator account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
	case errors.Is(err, errx.ErrorInitiatorInvalidSession):
		log.Info("initiator session is invalid")
		c.responser.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
	case errors.Is(err, errx.AccountHaveMembershipInOrg):
		c.responser.RenderErr(w, problems.Forbidden("account cannot be deleted while having membership in organization"))
	case err != nil:
		log.WithError(err).Error("failed to delete my account")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		log.Info("account deleted")
		c.responser.Render(w, http.StatusNoContent)
	}
}
