package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/restkit/ape"
	"github.com/netbill/restkit/ape/problems"
)

func (s *Service) DeleteMySessions(w http.ResponseWriter, r *http.Request) {
	initiator, err := rest.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get account from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get account from context"))

		return
	}

	if err = s.core.DeleteOwnSessions(r.Context(), account.InitiatorData{
		AccountID: initiator.AccountID,
		SessionID: initiator.SessionID,
	}); err != nil {
		s.log.WithError(err).Errorf("failed to delete My sessions")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			ape.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	ape.Render(w, http.StatusNoContent)
}
