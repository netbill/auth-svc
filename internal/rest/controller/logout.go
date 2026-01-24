package controller

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/restkit/ape"
	"github.com/netbill/restkit/ape/problems"
)

func (s *Service) Logout(w http.ResponseWriter, r *http.Request) {
	initiator, err := rest.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	err = s.core.Logout(r.Context(), account.InitiatorData{
		AccountID: initiator.AccountID,
		SessionID: initiator.SessionID,
	})
	if err != nil {
		s.log.WithError(err).Errorf("failed to logout user")
		switch {
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	ape.Render(w, http.StatusNoContent)
}
