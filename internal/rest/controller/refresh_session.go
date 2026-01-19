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

func (s *Service) RefreshSession(w http.ResponseWriter, r *http.Request) {
	req, err := requests.RefreshSession(r)
	if err != nil {
		s.log.WithError(err).Error("failed to parse refresh session request")
		ape.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	tokensPair, err := s.core.Refresh(r.Context(), req.Data.Attributes.RefreshToken)
	if err != nil {
		s.log.WithError(err).Errorf("failed to refresh session token")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("account not found"))
		case errors.Is(err, errx.ErrorInitiatorIsNotActive):
			ape.RenderErr(w, problems.Forbidden("account is not active"))
		case errors.Is(err, errx.ErrorSessionNotFound):
			ape.RenderErr(w, problems.Unauthorized("session not found"))
		case errors.Is(err, errx.ErrorSessionTokenMismatch):
			ape.RenderErr(w, problems.Forbidden("refresh session token mismatch"))
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	ape.Render(w, http.StatusOK, responses.TokensPair(tokensPair))
}
