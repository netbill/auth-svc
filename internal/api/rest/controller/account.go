package controller

import (
	"errors"
	"net/http"
	"slices"

	"github.com/netbill/auth-svc/internal/api/rest/responses"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/restkit"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationGetMyAccount = "get_my_account"

func (c *AccountController) GetMyAccount(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMyAccount)

	account, err := c.accounts.GetMyAccountByID(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountDeleted):
		log.WithError(err).Warn("account deleted")
		render.ResponseError(w, problems.Unauthorized())
		return
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.WithError(err).Warn("account not found")
		render.ResponseError(w, problems.NotFound("account not found"))
		return
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
		return
	}

	opts := make([]responses.AccountOption, 0, 1)
	includes := restkit.ParseIncludes(r)

	if slices.Contains(includes, "email") {
		email, err := c.accounts.GetMyEmailByID(r.Context(), scope.AccountActor(r))
		if err != nil {
			log.WithError(err).Error("failed to get account email")
		}

		opts = append(opts, responses.WithAccountEmail(email))
	}

	render.Response(w, http.StatusOK, responses.Account(account, opts...))
}

const operationDeleteMyAccount = "delete_my_account"

func (c *AccountController) DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMyAccount)

	err := c.accounts.DeleteMyAccount(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession),
		errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.WithError(err).Warn("account not found")
		render.ResponseError(w, problems.NotFound("account not found"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("account deleted")
		render.Response(w, http.StatusNoContent, nil)
	}
}
