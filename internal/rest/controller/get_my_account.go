package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
)

const operationGetMyAccount = "get_my_account"

func (c *Controller) GetMyAccount(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMyAccount)

	account, err := c.core.GetMyAccountByID(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound) || errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Infof("account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("account not found by credentials"))
	case err != nil:
		log.WithError(err).Error("failed to get my account")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		c.responser.Render(w, http.StatusOK, responses.Account(account))
	}
}
