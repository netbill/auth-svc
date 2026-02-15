package controller

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
)

const operationUpdatePassword = "update_password"

func (c *Controller) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdatePassword)

	req, err := requests.UpdatePassword(r)
	if err != nil {
		log.WithError(err).Info("invalid update password request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	err = c.core.UpdatePassword(
		r.Context(),
		scope.AccountActor(r),
		req.Data.Attributes.OldPassword,
		req.Data.Attributes.NewPassword,
	)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound) || errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Info("account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("failed to update password user not found"))
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.Info("invalid password")
		c.responser.RenderErr(w, problems.Unauthorized("invalid password"))
	case errors.Is(err, errx.ErrorCannotChangePasswordYet):
		log.Info("cannot change password yet")
		c.responser.RenderErr(w, problems.Forbidden("cannot change password yet"))
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.Info("password is not allowed")
		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"repo/attributes/password": err,
		})...)
	case err != nil:
		log.WithError(err).Error("update password failed")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		c.responser.Status(w, http.StatusNoContent)
	}
}
