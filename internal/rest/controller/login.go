package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationLoginByUsername = "login_by_username"

func (c *AuthController) LoginByUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationLoginByUsername)

	req, err := requests.LoginByUsername(r)
	if err != nil {
		log.WithError(err).Info("invalid login request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	log = log.WithField("username", req.Data.Attributes.Username)

	token, err := c.auth.LoginByUsername(r.Context(), req.Data.Attributes.Username, req.Data.Attributes.Password)
	switch {
	case errors.Is(err, errx.ErrorPasswordInvalid),
		errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountDeleted):
		log.WithError(err).Warn("invalid login or password")
		render.ResponseError(w, problems.Unauthorized())
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.TokensPair(token))
	}
}

const operationLoginByEmail = "login_by_email"

func (c *AuthController) LoginByEmail(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationLoginByEmail)

	req, err := requests.LoginByEmail(r)
	if err != nil {
		log.WithError(err).Info("invalid login request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	log = log.WithField("email", req.Data.Attributes.Email)

	token, err := c.auth.LoginByEmail(r.Context(), req.Data.Attributes.Email, req.Data.Attributes.Password)
	switch {
	case errors.Is(err, errx.ErrorPasswordInvalid),
		errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountDeleted):
		log.WithError(err).Warn("invalid login or password")
		render.ResponseError(w, problems.Unauthorized())
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.TokensPair(token))
	}
}
