package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const operationUpdateUsername = "update_username"

func (c *Controller) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdateUsername)

	req, err := requests.UpdateUsername(r)
	if err != nil {
		log.WithError(err).Info("invalid update username request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	res, err := c.core.UpdateUsername(r.Context(), scope.AccountActor(r), req.Data.Attributes.NewUsername)
	switch {
	case errors.Is(err, errx.ErrorInitiatorNotFound):
		log.Info("initiator account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("failed to update username user not found"))
	case errors.Is(err, errx.ErrorInitiatorInvalidSession):
		log.Info("initiator session is invalid")
		c.responser.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.Info("invalid password")
		c.responser.RenderErr(w, problems.Unauthorized("invalid password"))
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.Info("username already taken")
		c.responser.RenderErr(w, problems.Conflict("user with this username already exists"))
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.Info("username is not allowed")
		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"data/attributes/new_username": err,
		})...)
	case err != nil:
		log.WithError(err).Error("update username failed")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		c.responser.Render(w, http.StatusOK, responses.Account(res))
	}
}
