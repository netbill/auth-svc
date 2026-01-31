package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/contexter"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/problems"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (c *Controller) RegistrationByAdmin(w http.ResponseWriter, r *http.Request) {
	initiator, err := contexter.AccountData(r.Context())
	if err != nil {
		c.log.WithError(err).Error("failed to get user from context")
		c.responser.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	req, err := requests.RegistrationAdmin(r)
	if err != nil {
		c.log.WithError(err).Error("failed to decode register admin request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	u, err := c.core.Registration(
		r.Context(),
		account.RegistrationParams{
			Email:    req.Data.Attributes.Email,
			Username: req.Data.Attributes.Username,
			Password: req.Data.Attributes.Password,
			Role:     req.Data.Attributes.Role,
		},
	)
	if err != nil {
		c.log.WithError(err).Errorf("failed to register by admin")
		switch {
		case errors.Is(err, errx.ErrorNotEnoughRights):
			c.responser.RenderErr(w, problems.Forbidden("only admins can register new admin accounts"))
		case errors.Is(err, errx.ErrorEmailAlreadyExist):
			c.responser.RenderErr(w, problems.Conflict("user with this email already exists"))
		case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
			c.responser.RenderErr(w, problems.Conflict("user with this username already exists"))
		case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
			c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/username": err,
			})...)
		case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
			c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/password": err,
			})...)
		case errors.Is(err, errx.ErrorRoleNotSupported):
			c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/role": err,
			})...)
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.log.Infof("admin %s registered successfully by user %s", u.ID, initiator)

	c.responser.Render(w, http.StatusCreated, responses.Account(u))
}
