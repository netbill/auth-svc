package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/restkit/ape"
	"github.com/netbill/restkit/ape/problems"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (s *Service) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	initiator, err := rest.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	req, err := requests.UpdatePassword(r)
	if err != nil {
		s.log.WithError(err).Error("failed to decode update password request")
		ape.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	err = s.core.UpdatePassword(r.Context(), account.InitiatorData{
		AccountID: initiator.AccountID,
		SessionID: initiator.SessionID,
	}, req.Data.Attributes.OldPassword, req.Data.Attributes.NewPassword)
	if err != nil {
		s.log.WithError(err).Errorf("failed to update password")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("failed to update password user not found"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			ape.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		case errors.Is(err, errx.ErrorPasswordInvalid):
			ape.RenderErr(w, problems.Unauthorized("invalid password"))
		case errors.Is(err, errx.ErrorCannotChangePasswordYet):
			ape.RenderErr(w, problems.Forbidden("cannot change password yet"))
		case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
			ape.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/password": err,
			})...)
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	ape.Render(w, http.StatusNoContent)
}
