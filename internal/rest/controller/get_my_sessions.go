package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/ape/problems"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/middlewares"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/pagi"
)

func (s *Service) GetMySessions(w http.ResponseWriter, r *http.Request) {
	initiator, err := middlewares.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	limit, offset := pagi.GetPagination(r)
	sessions, err := s.core.GetOwnSessions(r.Context(), account.InitiatorData{
		AccountID: initiator.AccountID,
		SessionID: initiator.SessionID,
	}, limit, offset)
	if err != nil {
		s.log.WithError(err).Errorf("failed to select My sessions")
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

	ape.Render(w, http.StatusOK, responses.AccountSessionsCollection(r, sessions))
}
