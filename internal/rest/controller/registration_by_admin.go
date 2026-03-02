package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/auth"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const operationRegistrationByAdmin = "registration_by_admin"

func (c *Controller) RegistrationByAdmin(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRegistrationByAdmin)

	req, err := requests.RegistrationAdmin(r)
	if err != nil {
		log.WithError(err).Info("invalid registration admin request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	log = log.WithFields(map[string]interface{}{
		"email":    req.Data.Attributes.Email,
		"username": req.Data.Attributes.Username,
		"role":     req.Data.Attributes.Role,
	})

	u, err := c.core.Registration(
		r.Context(),
		auth.RegistrationParams{
			Email:    req.Data.Attributes.Email,
			Username: req.Data.Attributes.Username,
			Password: req.Data.Attributes.Password,
			Role:     req.Data.Attributes.Role,
		},
	)
	switch {
	case errors.Is(err, errx.ErrorNotEnoughRights):
		log.WithError(err).Warn("not enough rights to register admin")
		render.ResponseError(w, problems.Forbidden("only admins can register new admin accounts"))
	case errors.Is(err, errx.ErrorEmailAlreadyExist):
		log.WithError(err).Warn("email already exists")
		render.ResponseError(w, problems.Conflict("user with this email already exists"))
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.WithError(err).Warn("username already taken")
		render.ResponseError(w, problems.Conflict("user with this username already exists"))
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.WithError(err).Warn("username is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/username": err,
		})...)
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.WithError(err).Warn("password is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/password": err,
		})...)
	case errors.Is(err, errx.ErrorRoleNotSupported):
		log.WithError(err).Warn("role is not supported")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/role": err,
		})...)
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("account registered successfully by admin")
		render.Response(w, http.StatusCreated, responses.Account(u))
	}
}
