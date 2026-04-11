package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/netbill/auth-svc/internal/api/rest/requests"
	"github.com/netbill/auth-svc/internal/api/rest/responses"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
	"golang.org/x/oauth2"
)

const operationLoginByUsername = "login_by_username"

func (c *SessionController) LoginByUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationLoginByUsername)

	req, err := requests.LoginByUsername(r)
	if err != nil {
		log.WithError(err).Info("invalid login request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	log = log.WithField("username", req.Data.Attributes.Username)

	token, err := c.sessions.LoginByUsername(r.Context(), req.Data.Attributes.Username, req.Data.Attributes.Password)
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

func (c *SessionController) LoginByEmail(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationLoginByEmail)

	req, err := requests.LoginByEmail(r)
	if err != nil {
		log.WithError(err).Info("invalid login request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	log = log.WithField("email", req.Data.Attributes.Email)

	token, err := c.sessions.LoginByEmail(r.Context(), req.Data.Attributes.Email, req.Data.Attributes.Password)
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

func (c *SessionController) LoginByGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	url := c.google.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

const operationLoginByGoogleOAuthCallback = "login_by_google_oauth"

func (c *SessionController) LoginByGoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationLoginByGoogleOAuthCallback)

	code := r.URL.Query().Get("code")
	if code == "" {
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"query": fmt.Errorf("code is required"),
		})...)
		return
	}

	token, err := c.google.Exchange(r.Context(), code)
	if err != nil {
		log.WithError(err).Error("google oauth exchange failed")
		render.ResponseError(w, problems.InternalError())
		return
	}

	client := c.google.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.WithError(err).Error("google userinfo request failed")
		render.ResponseError(w, problems.InternalError())
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Warn("failed to close google userinfo response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.WithField("google_status", resp.StatusCode).Error("google userinfo returned non-200")
		render.ResponseError(w, problems.InternalError())
		return
	}

	var userInfo struct {
		Email string `json:"email"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.WithError(err).Error("failed to decode google userinfo")
		render.ResponseError(w, problems.InternalError())
		return
	}

	log = log.WithField("user_email", userInfo.Email)

	tokensPair, err := c.sessions.LoginByGoogle(r.Context(), userInfo.Email)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountDeleted):
		log.WithError(err).Warn("account with this email not found")
		render.ResponseError(w, problems.NotFound("user with this email not found"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("login by google successful")
		render.Response(w, http.StatusOK, responses.TokensPair(tokensPair))
	}
}
