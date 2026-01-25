package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/ape/problems"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/middlewares"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (s *Service) RegistrationByAdmin(w http.ResponseWriter, r *http.Request) {
	initiator, err := middlewares.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	req, err := requests.RegistrationAdmin(r)
	if err != nil {
		s.log.WithError(err).Error("failed to decode register admin request")
		ape.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	u, err := s.core.RegistrationByAdmin(r.Context(),
		initiator.AccountID,
		account.RegistrationParams{
			Email:    req.Data.Attributes.Email,
			Username: req.Data.Attributes.Username,
			Password: req.Data.Attributes.Password,
			Role:     req.Data.Attributes.Role,
		},
	)
	if err != nil {
		s.log.WithError(err).Errorf("failed to register by admin")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			ape.RenderErr(w, problems.Unauthorized("failed to register admin user not found"))
		case errors.Is(err, errx.ErrorNotEnoughRights):
			ape.RenderErr(w, problems.Forbidden("only admins can register new admin accounts"))
		case errors.Is(err, errx.ErrorEmailAlreadyExist):
			ape.RenderErr(w, problems.Conflict("user with this email already exists"))
		case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
			ape.RenderErr(w, problems.Conflict("user with this username already exists"))
		case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
			ape.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/username": err,
			})...)
		case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
			ape.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/password": err,
			})...)
		case errors.Is(err, errx.ErrorRoleNotSupported):
			ape.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/role": err,
			})...)
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	s.log.Infof("admin %s registered successfully by user %s", u.ID, initiator)

	ape.Render(w, http.StatusCreated, responses.Account(u))
}
