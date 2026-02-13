package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
)

const operationLoginByUsername = "login_by_username"

func (c *Controller) LoginByUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationLoginByUsername)

	req, err := requests.LoginByUsername(r)
	if err != nil {
		log.WithError(err).Info("invalid login request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	token, err := c.core.LoginByUsername(r.Context(), req.Data.Attributes.Username, req.Data.Attributes.Password)
	switch {
	case errors.Is(err, errx.ErrorPasswordInvalid) || errors.Is(err, errx.ErrorAccountNotFound):
		log.Info("invalid login or password")
		c.responser.RenderErr(w, problems.Unauthorized("invalid login or password"))
	case err != nil:
		log.WithError(err).Error("login by username failed")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		c.responser.Render(w, http.StatusOK, responses.TokensPair(token))
	}
}
