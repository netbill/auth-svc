package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/contexter"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/restkit/problems"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (c *Controller) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	initiator, err := contexter.AccountData(r.Context())
	if err != nil {
		c.log.WithError(err).Error("failed to get user from context")
		c.responser.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	req, err := requests.UpdatePassword(r)
	if err != nil {
		c.log.WithError(err).Error("failed to decode update password request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	err = c.core.UpdatePassword(r.Context(), account.InitiatorData{
		AccountID: initiator.GetAccountID(),
		SessionID: initiator.GetSessionID(),
	}, req.Data.Attributes.OldPassword, req.Data.Attributes.NewPassword)
	if err != nil {
		c.log.WithError(err).Errorf("failed to update password")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("failed to update password user not found"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			c.responser.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		case errors.Is(err, errx.ErrorPasswordInvalid):
			c.responser.RenderErr(w, problems.Unauthorized("invalid password"))
		case errors.Is(err, errx.ErrorCannotChangePasswordYet):
			c.responser.RenderErr(w, problems.Forbidden("cannot change password yet"))
		case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
			c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/password": err,
			})...)
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.responser.Render(w, http.StatusNoContent)
}
