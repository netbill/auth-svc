package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/api/rest/requests"
	"github.com/netbill/auth-svc/internal/api/rest/responses"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationUpdateUsername = "update_username"

func (c *AccountController) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdateUsername)

	req, err := requests.UpdateUsername(r)
	if err != nil {
		log.WithError(err).Info("failed to parse update username request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	res, err := c.accounts.UpdateUsername(r.Context(), scope.AccountActor(r), req.Data.Attributes.Username)
	switch {
	case errors.Is(err, errx.ErrorAccountDeleted):
		log.WithError(err).Warn("account deleted")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.WithError(err).Warn("account not found")
		render.ResponseError(w, problems.NotFound("account not found"))
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.WithError(err).Warn("username is already taken")
		render.ResponseError(w, problems.Conflict("user with this username already exists"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("username updated successfully")
		render.Response(w, http.StatusOK, responses.Account(res))
	}
}
