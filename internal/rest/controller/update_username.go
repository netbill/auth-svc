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
	"github.com/netbill/auth-svc/internal/rest/responses"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (s *Service) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	initiator, err := meta.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	req, err := requests.UpdateUsername(r)
	if err != nil {
		s.log.WithError(err).Error("failed to decode update username request")
		ape.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	res, err := s.domain.UpdateUsername(r.Context(), auth.InitiatorData{
		AccountID: initiator.ID,
		SessionID: initiator.SessionID,
	}, req.Data.Attributes.Password, req.Data.Attributes.NewUsername)
	if err != nil {
		s.log.WithError(err).Errorf("failed to update username")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("failed to update password user not found"))
		case errors.Is(err, errx.ErrorInitiatorIsNotActive):
			ape.RenderErr(w, problems.Forbidden("initiator is blocked"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			ape.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		case errors.Is(err, errx.ErrorPasswordInvalid):
			ape.RenderErr(w, problems.Unauthorized("invalid password"))
		case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
			ape.RenderErr(w, problems.Conflict("user with this username already exists"))
		case errors.Is(err, errx.ErrorCannotChangeUsernameYet):
			ape.RenderErr(w, problems.Forbidden("cannot change username due to cooldown"))
		case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
			ape.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/new_username": err,
			})...)
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	ape.Render(w, http.StatusOK, responses.Account(res))
}
