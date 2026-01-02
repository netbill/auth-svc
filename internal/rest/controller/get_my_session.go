package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/ape/problems"
	"github.com/netbill/auth-svc/internal/domain/errx"
	"github.com/netbill/auth-svc/internal/domain/modules/auth"
	"github.com/netbill/auth-svc/internal/rest/meta"
	"github.com/netbill/auth-svc/internal/rest/responses"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

func (s *Service) GetMySession(w http.ResponseWriter, r *http.Request) {
	initiator, err := meta.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	sessionId, err := uuid.Parse(chi.URLParam(r, "session_id"))
	if err != nil {
		s.log.WithError(err).Errorf("invalid session id: %s", chi.URLParam(r, "session_id"))
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"query": fmt.Errorf("invalid session id: %s", chi.URLParam(r, "session_id")),
		})...)

		return
	}

	session, err := s.domain.GetOwnSession(r.Context(), auth.InitiatorData{
		AccountID: initiator.ID,
		SessionID: initiator.SessionID,
	}, sessionId)
	if err != nil {
		s.log.WithError(err).Errorf("failed to get My session")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		case errors.Is(err, errx.ErrorInitiatorIsNotActive):
			ape.RenderErr(w, problems.Forbidden("initiator is blocked"))
		case errors.Is(err, errx.ErrorSessionNotFound):
			ape.RenderErr(w, problems.Unauthorized("session not found"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			ape.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	ape.Render(w, http.StatusOK, responses.AccountSession(session))
}
