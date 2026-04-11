package controller

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/netbill/auth-svc/internal/api/rest/requests"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationUpdatePassword = "update_password"

func (c *AccountController) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdatePassword)

	req, err := requests.UpdatePassword(r)
	if err != nil {
		log.WithError(err).Info("failed to parse update password request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	err = c.accounts.UpdatePassword(
		r.Context(),
		scope.AccountActor(r),
		req.Data.Attributes.OldPassword,
		req.Data.Attributes.NewPassword,
	)
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession),
		errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.WithError(err).Warn("invalid old password")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.WithError(err).Warn("new password is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/new_password": err,
		})...)
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("password updated successfully")
		render.Response(w, http.StatusNoContent, nil)
	}
}
