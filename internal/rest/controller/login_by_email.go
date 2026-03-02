package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationLoginByEmail = "login_by_email"

func (c *Controller) LoginByEmail(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationLoginByEmail)

	req, err := requests.LoginByEmail(r)
	if err != nil {
		log.WithError(err).Info("invalid login request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	log = log.WithField("email", req.Data.Attributes.Email)

	token, err := c.core.LoginByEmail(r.Context(), req.Data.Attributes.Email, req.Data.Attributes.Password)
	switch {
	case errors.Is(err, errx.ErrorPasswordInvalid),
		errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountDeleted):
		log.WithError(err).Warn("invalid login or password")
		render.ResponseError(w, problems.Unauthorized("invalid login or password"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.TokensPair(token))
	}
}
