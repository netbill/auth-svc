package controller

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationUpdateUsername = "update_username"

func (c *AuthController) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdateUsername)

	req, err := requests.UpdateUsername(r)
	if err != nil {
		log.WithError(err).Info("failed to parse update username request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	res, err := c.auth.UpdateUsername(r.Context(), scope.AccountActor(r), req.Data.Attributes.Username)
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.WithError(err).Warn("invalid password")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.WithError(err).Warn("username is already taken")
		render.ResponseError(w, problems.Conflict("user with this username already exists"))
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.WithError(err).Warn("username is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/username": err,
		})...)
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("username updated successfully")
		render.Response(w, http.StatusOK, responses.Account(res))
	}
}
