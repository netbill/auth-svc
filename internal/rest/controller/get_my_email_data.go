package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/ape"
	"github.com/netbill/restkit/ape/problems"
)

func (s *Service) GetMyEmailData(w http.ResponseWriter, r *http.Request) {
	initiator, err := rest.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	emailData, err := s.core.GetAccountEmail(r.Context(), initiator.AccountID)
	if err != nil {
		s.log.WithError(err).Errorf("failed to get email repo by id: %s", initiator.AccountID)
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	ape.Render(w, http.StatusOK, responses.AccountEmailData(emailData))
}
