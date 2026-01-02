package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/ape/problems"
	"github.com/netbill/auth-svc/internal/domain/errx"
	"github.com/netbill/auth-svc/internal/domain/modules/auth"
	"github.com/netbill/auth-svc/internal/rest/meta"
)

func (s *Service) DeleteMySessions(w http.ResponseWriter, r *http.Request) {
	initiator, err := meta.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get account from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get account from context"))

		return
	}

	if err = s.domain.DeleteOwnSessions(r.Context(), auth.InitiatorData{
		AccountID: initiator.ID,
		SessionID: initiator.SessionID,
	}); err != nil {
		s.log.WithError(err).Errorf("failed to delete My sessions")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		case errors.Is(err, errx.ErrorInitiatorIsNotActive):
			ape.RenderErr(w, problems.Forbidden("initiator is not active"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			ape.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
