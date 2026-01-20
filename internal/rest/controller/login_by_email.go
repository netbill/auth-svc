package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/ape"
	"github.com/netbill/restkit/ape/problems"
)

func (s *Service) LoginByEmail(w http.ResponseWriter, r *http.Request) {
	req, err := requests.LoginByEmail(r)
	if err != nil {
		s.log.WithError(err).Error("failed to decode login request")
		ape.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	token, err := s.core.LoginByEmail(r.Context(), req.Data.Attributes.Email, req.Data.Attributes.Password)
	if err != nil {
		s.log.WithError(err).Errorf("failed to login user")
		switch {
		case errors.Is(err, errx.ErrorPasswordInvalid) || errors.Is(err, errx.ErrorAccountNotFound):
			ape.RenderErr(w, problems.Unauthorized("invalid login or password"))
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	s.log.Infof("user %s logged in successfully", req.Data.Attributes.Email)

	ape.Render(w, http.StatusOK, responses.TokensPair(token))
}
