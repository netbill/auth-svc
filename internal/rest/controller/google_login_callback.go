package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/ape"
	"github.com/netbill/restkit/ape/problems"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (s *Service) LoginByGoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"query": fmt.Errorf("code is required"),
		})...)

		return
	}

	token, err := s.google.Exchange(r.Context(), code)
	if err != nil {
		s.log.WithError(err).Errorf("error exchanging code for user id: %s", code)
		ape.RenderErr(w, problems.InternalError())

		return
	}

	client := s.google.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		s.log.WithError(err).Errorf("error getting user info from Google")
		ape.RenderErr(w, problems.InternalError())

		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			s.log.WithError(err).Errorf("error closing response body")
			ape.RenderErr(w, problems.InternalError())

			return
		}
	}(resp.Body)

	var userInfo struct {
		Email string `json:"email"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		s.log.WithError(err).Errorf("error decoding user info from Google")
		ape.RenderErr(w, problems.InternalError())

		return
	}

	tokensPair, err := s.core.LoginByGoogle(r.Context(), userInfo.Email)
	if err != nil {
		s.log.WithError(err).Errorf("error logging in user: %s", userInfo.Email)
		switch {
		case errors.Is(err, errx.ErrorInitiatorIsNotActive):
			ape.RenderErr(w, problems.Forbidden("account is not active"))
		case errors.Is(err, errx.ErrorAccountNotFound):
			ape.RenderErr(w, problems.NotFound("user with this email not found"))
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return

	}

	s.log.Infof("Account %s logged in with Google", userInfo.Email)

	ape.Render(w, http.StatusOK, responses.TokensPair(tokensPair))
}
