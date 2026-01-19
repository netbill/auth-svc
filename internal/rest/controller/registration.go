package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape/problems"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/restkit/ape"
	"github.com/netbill/restkit/auth/roles"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (s *Service) Registration(w http.ResponseWriter, r *http.Request) {
	req, err := requests.Registration(r)
	if err != nil {
		s.log.WithError(err).Error("failed to decode register request")
		ape.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	_, err = s.core.Registration(r.Context(), account.RegistrationParams{
		Email:    req.Data.Attributes.Email,
		Password: req.Data.Attributes.Password,
		Role:     roles.SystemUser,
	})
	if err != nil {
		s.log.WithError(err).Errorf("failed to register user")
		switch {
		case errors.Is(err, errx.ErrorEmailAlreadyExist):
			ape.RenderErr(w, problems.Conflict("user with this email already exists"))
		case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
			ape.RenderErr(w, problems.BadRequest(validation.Errors{
				"repo/attributes/password": err,
			})...)
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	s.log.Infof("user %s registered successfully", req.Data.Attributes.Email)

	w.WriteHeader(http.StatusCreated)
}
