package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/contexter"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/problems"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (c *Controller) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	initiator, err := contexter.AccountData(r.Context())
	if err != nil {
		c.log.WithError(err).Error("failed to get user from context")
		c.responser.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	req, err := requests.UpdateUsername(r)
	if err != nil {
		c.log.WithError(err).Error("failed to decode update username request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	res, err := c.core.UpdateUsername(r.Context(), account.InitiatorData{
		AccountID: initiator.GetAccountID(),
		SessionID: initiator.GetSessionID(),
	}, req.Data.Attributes.NewUsername)
	if err != nil {
		c.log.WithError(err).Errorf("failed to update username")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("failed to update password user not found"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			c.responser.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		case errors.Is(err, errx.ErrorPasswordInvalid):
			c.responser.RenderErr(w, problems.Unauthorized("invalid password"))
		case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
			c.responser.RenderErr(w, problems.Conflict("user with this username already exists"))
		case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
			c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
				"data/attributes/new_username": err,
			})...)
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.responser.Render(w, http.StatusOK, responses.Account(res))
}
