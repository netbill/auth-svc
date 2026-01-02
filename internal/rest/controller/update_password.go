package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/ape/problems"
	"github.com/netbill/auth-svc/internal/domain/errx"
	"github.com/netbill/auth-svc/internal/domain/modules/auth"
	"github.com/netbill/auth-svc/internal/rest/meta"
	"github.com/netbill/auth-svc/internal/rest/requests"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (s *Service) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	initiator, err := meta.AccountData(r.Context())
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

	err = s.domain.UpdatePassword(r.Context(), auth.InitiatorData{
		AccountID: initiator.ID,
		SessionID: initiator.SessionID,
	}, req.Data.Attributes.OldPassword, req.Data.Attributes.NewPassword)
	if err != nil {
		s.log.WithError(err).Errorf("failed to update password")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("failed to update password user not found"))
		case errors.Is(err, errx.ErrorInitiatorIsNotActive):
			ape.RenderErr(w, problems.Forbidden("initiator is blocked"))
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

	w.WriteHeader(http.StatusNoContent)
}
