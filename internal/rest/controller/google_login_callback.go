package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/problems"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (c *Controller) LoginByGoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"query": fmt.Errorf("code is required"),
		})...)

		return
	}

	token, err := c.google.Exchange(r.Context(), code)
	if err != nil {
		c.log.WithError(err).Errorf("error exchanging code for user id: %s", code)
		c.responser.RenderErr(w, problems.InternalError())

		return
	}

	client := c.google.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.log.WithError(err).Errorf("error getting user info from Google")
		c.responser.RenderErr(w, problems.InternalError())

		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			c.log.WithError(err).Errorf("error closing response body")
			c.responser.RenderErr(w, problems.InternalError())

			return
		}
	}(resp.Body)

	var userInfo struct {
		Email string `json:"email"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.log.WithError(err).Errorf("error decoding user info from Google")
		c.responser.RenderErr(w, problems.InternalError())

		return
	}

	tokensPair, err := c.core.LoginByGoogle(r.Context(), userInfo.Email)
	if err != nil {
		c.log.WithError(err).Errorf("error logging in user: %s", userInfo.Email)
		switch {
		case errors.Is(err, errx.ErrorAccountNotFound):
			c.responser.RenderErr(w, problems.NotFound("user with this email not found"))
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return

	}

	c.log.Infof("Account %s logged in with Google", userInfo.Email)

	c.responser.Render(w, http.StatusOK, responses.TokensPair(tokensPair))
}
