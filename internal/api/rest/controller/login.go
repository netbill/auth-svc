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

	defer c.metrics.RecordUsernameLogin(r.Context(), &err)
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
		log.Info("login by username successful")
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

	defer c.metrics.RecordEmailLogin(r.Context(), &err)
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
		log.Info("login by email successful")
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

	defer c.metrics.RecordGoogleLogin(r.Context(), &err)
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

func (c *SessionController) QRConnect(w http.ResponseWriter, r *http.Request) {
	c.wsHandler.LoginWithQR(w, r, nil)
}

const operationQRConfirm = "qr_confirm"

func (c *SessionController) QRConfirm(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationQRConfirm)

	req, err := requests.QRConfirm(r)
	if err != nil {
		log.WithError(err).Warn("invalid qr confirm request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	qrToken := req.Data.Attributes.QrToken.String()

	defer c.metrics.RecordQRLogin(r.Context(), &err)

	pair, err := c.sessions.ConfirmQRToken(r.Context(), scope.AccountActor(r), qrToken)
	switch {
	case errors.Is(err, errx.ErrorQRTokenNotFound):
		log.WithError(err).Warn("qr token not found")
		render.ResponseError(w, problems.NotFound("qr token not found or expired"))
		return
	case errors.Is(err, errx.ErrorQRTokenAlreadyConfirmed):
		log.WithError(err).Warn("qr token already confirmed")
		render.ResponseError(w, problems.Conflict("qr token already confirmed"))
		return
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
		return
	}

	payload, err := json.Marshal(pair)
	if err != nil {
		log.WithError(err).Error("failed to marshal tokens pair")
		render.ResponseError(w, problems.InternalError())
		return
	}

	if err = c.sessions.PublishQRToken(r.Context(), qrToken, payload); err != nil {
		log.WithError(err).Error("failed to publish qr confirm")
		render.ResponseError(w, problems.InternalError())
		return
	}

	log.Info("qr token confirmed")
	render.Response(w, http.StatusNoContent, nil)
}
