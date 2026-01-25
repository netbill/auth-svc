package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/ape/problems"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/middlewares"
)

func (s *Service) DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	initiator, err := middlewares.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	err = s.core.DeleteOwnAccount(r.Context(), account.InitiatorData{
		AccountID: initiator.AccountID,
		SessionID: initiator.SessionID,
	})
	if err != nil {
		s.log.WithError(err).Errorf("failed to delete my account with id: %s", initiator.AccountID)
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			ape.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		case errors.Is(err, errx.AccountHaveMembershipInOrg):
			ape.RenderErr(w, problems.Forbidden("account cannot be deleted while having membership in organization"))
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	ape.Render(w, http.StatusNoContent)
}
