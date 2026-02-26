package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationDeleteMyAccount = "delete_my_account"

func (c *Controller) DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMyAccount)

	err := c.core.DeleteMyAccount(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound) || errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Info("account not found by credentials")
		render.ResponseError(w, problems.Unauthorized("account not found by credentials"))
	case errors.Is(err, errx.AccountHaveMembershipInOrg):
		render.ResponseError(w, problems.Forbidden("account cannot be deleted while having membership in organization"))
	case err != nil:
		log.WithError(err).Error("failed to delete my account")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("account deleted")
		render.Response(w, http.StatusNoContent, nil)
	}
}
