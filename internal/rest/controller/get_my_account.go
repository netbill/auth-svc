package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/ape/problems"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/rest/responses"
)

func (s *Service) GetMyAccount(w http.ResponseWriter, r *http.Request) {
	initiator, err := rest.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get account from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get account from context"))

		return
	}

	account, err := s.core.GetAccountByID(r.Context(), initiator.AccountID)
	if err != nil {
		s.log.WithError(err).Errorf("failed to get account by id: %s", initiator.AccountID)
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	ape.Render(w, http.StatusOK, responses.Account(account))
}
