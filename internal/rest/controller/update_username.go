package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const operationUpdateUsername = "update_username"

func (c *Controller) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdateUsername)

	req, err := requests.UpdateUsername(r)
	if err != nil {
		log.WithError(err).Info("invalid update username request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	res, err := c.core.UpdateUsername(r.Context(), scope.AccountActor(r), req.Data.Attributes.Username)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound) || errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Info("invalid credentials")
		render.ResponseError(w, problems.Unauthorized("failed to update password user not found"))
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.Info("invalid password")
		render.ResponseError(w, problems.Unauthorized("invalid password"))
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.Info("username already taken")
		render.ResponseError(w, problems.Conflict("user with this username already exists"))
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.Info("username is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/username": err,
		})...)
	case err != nil:
		log.WithError(err).Error("update username failed")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.Account(res))
	}
}
