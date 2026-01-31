package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/tokens"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (c *Controller) Registration(w http.ResponseWriter, r *http.Request) {
	req, err := requests.Registration(r)
	if err != nil {
		c.log.WithError(err).Error("failed to decode register request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	_, err = c.core.Registration(r.Context(), account.RegistrationParams{
		Email:    req.Data.Attributes.Email,
		Password: req.Data.Attributes.Password,
		Username: req.Data.Attributes.Username,
		Role:     tokens.RoleSystemAdmin,
	})
	if err != nil {
		c.log.WithError(err).Errorf("failed to register user")
		switch {
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
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.log.Infof("user %s registered successfully", req.Data.Attributes.Email)

	w.WriteHeader(http.StatusCreated)
}
