package controller

import (
	"errors"
	"net/http"
	"slices"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationGetMyAccount = "get_my_account"

func (c *AuthController) GetMyAccount(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMyAccount)

	account, err := c.auth.GetMyAccountByID(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
		return
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
		return
	}

	opts := make([]responses.AccountOption, 0, 1)
	includes := restkit.ParseIncludes(r)

	if slices.Contains(includes, "email") {
		email, err := c.auth.GetMyAccountEmail(r.Context(), scope.AccountActor(r))
		if err != nil {
			log.WithError(err).Error("failed to get account email")
		}

		opts = append(opts, responses.WithAccountEmail(email))
	}

	render.Response(w, http.StatusOK, responses.Account(account, opts...))
}

const operationDeleteMyAccount = "delete_my_account"

func (c *AuthController) DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMyAccount)

	err := c.auth.DeleteMyAccount(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorAccountHaveMembershipInOrg):
		log.WithError(err).Warn("account having membership in organization")
		render.ResponseError(w, problems.Forbidden("account cannot be deleted while having membership in organization"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("account deleted")
		render.Response(w, http.StatusNoContent, nil)
	}
}
