package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/problems"
)

func (c *Controller) LoginByUsername(w http.ResponseWriter, r *http.Request) {
	req, err := requests.LoginByUsername(r)
	if err != nil {
		c.log.WithError(err).Error("failed to decode login request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	token, err := c.core.LoginByUsername(r.Context(), req.Data.Attributes.Username, req.Data.Attributes.Password)
	if err != nil {
		c.log.WithError(err).Errorf("failed to login user")
		switch {
		case errors.Is(err, errx.ErrorPasswordInvalid) || errors.Is(err, errx.ErrorAccountNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("invalid login or password"))
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.log.Infof("user %s logged in successfully", req.Data.Attributes.Username)

	c.responser.Render(w, http.StatusOK, responses.TokensPair(token))
}
