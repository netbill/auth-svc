package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const operationLoginByGoogleOAuthCallback = "login_by_google_oauth_callback"

func (c *Controller) LoginByGoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
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

	tokensPair, err := c.core.LoginByGoogle(r.Context(), userInfo.Email)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.Info("account not found for google email")
		render.ResponseError(w, problems.NotFound("user with this email not found"))
	case err != nil:
		log.WithError(err).Error("login by google failed")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.TokensPair(tokensPair))
	}
}
