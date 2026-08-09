package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/api/rest/requests"
	"github.com/netbill/auth-svc/internal/api/rest/responses"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/restkit"
	"github.com/netbill/restkit/pagi"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
	"github.com/netbill/restkit/tokens"
)

type userCore interface {
	Registration(ctx context.Context, params user.RegistrationParams) (models.User, error)

	GetMyUserByID(ctx context.Context, actor models.UserActor) (models.User, error)
	GetMyEmailByID(ctx context.Context, actor models.UserActor) (models.UserEmail, error)

	GetUserByID(ctx context.Context, userID uuid.UUID) (models.User, error)
	GetByUsername(ctx context.Context, username string) (models.User, error)
	GetList(ctx context.Context, params user.FilterParams, limit, offset uint) (pagi.Page[[]models.User], error)

	Update(ctx context.Context, actor models.UserActor, params user.UpdateParams) (models.User, error)
	UpdateUsername(ctx context.Context, actor models.UserActor, username string) (models.User, error)

	UpdatePassword(ctx context.Context, actor models.UserActor, oldPassword, newPassword string) error

	CreateUploadMediaLinks(ctx context.Context, actor models.UserActor) (models.User, models.UploadUserMediaLinks, error)
	DeleteUploadMedia(ctx context.Context, actor models.UserActor, params user.DeleteUploadMediaParams) error

	DeleteMyUser(ctx context.Context, actor models.UserActor) error
}

type userMetrics interface {
	RecordRegistration(ctx context.Context, err *error)
}

type UserController struct {
	users   userCore
	metrics userMetrics
}

func NewUserController(users userCore, m userMetrics) *UserController {
	return &UserController{users: users, metrics: m}
}

const operationRegistration = "registration"

func (c *UserController) Registration(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRegistration)

	req, err := requests.Registration(r)
	if err != nil {
		log.WithError(err).Info("invalid registration request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	defer c.metrics.RecordRegistration(r.Context(), &err)
	_, err = c.users.Registration(r.Context(), user.RegistrationParams{
		Email:    req.Data.Attributes.Email,
		Password: req.Data.Attributes.Password,
		Username: req.Data.Attributes.Username,
		Role:     tokens.RoleSystemUser,
	})
	switch {
	case errors.Is(err, errx.ErrorEmailAlreadyExist):
		log.WithError(err).Warn("email already exists")
		render.ResponseError(w, problems.Conflict("user with this email already exists"))
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.WithError(err).Warn("password is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/password": err,
		})...)
	case errors.Is(err, errx.ErrorUsernameNotValid):
		log.WithError(err).Warn("username is invalid")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/username": err,
		})...)
	case errors.Is(err, errx.ErrorUsernameTaken):
		log.WithError(err).Warn("username is already taken")
		render.ResponseError(w, problems.Conflict("username is already taken"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("registration successful")
		w.WriteHeader(http.StatusCreated)
	}
}

const operationRegistrationByAdmin = "registration_by_admin"

func (c *UserController) RegistrationByAdmin(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRegistrationByAdmin)

	req, err := requests.RegistrationAdmin(r)
	if err != nil {
		log.WithError(err).Info("invalid registration admin request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	log = log.WithFields(map[string]interface{}{
		"email": req.Data.Attributes.Email,
		"role":  req.Data.Attributes.Role,
	})

	defer c.metrics.RecordRegistration(r.Context(), &err)
	u, err := c.users.Registration(r.Context(), user.RegistrationParams{
		Email:    req.Data.Attributes.Email,
		Password: req.Data.Attributes.Password,
		Username: req.Data.Attributes.Username,
		Role:     req.Data.Attributes.Role,
	})
	switch {
	case errors.Is(err, errx.ErrorEmailAlreadyExist):
		log.WithError(err).Warn("email already exists")
		render.ResponseError(w, problems.Conflict("user with this email already exists"))
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.WithError(err).Warn("password is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/password": err,
		})...)
	case errors.Is(err, errx.ErrorUsernameNotValid):
		log.WithError(err).Warn("username is invalid")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/username": err,
		})...)
	case errors.Is(err, errx.ErrorUsernameTaken):
		log.WithError(err).Warn("username is already taken")
		render.ResponseError(w, problems.Conflict("username is already taken"))
	case errors.Is(err, errx.ErrorRoleNotSupported):
		log.WithError(err).Warn("role is not supported")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/role": err,
		})...)
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("user registered successfully by admin")
		render.Response(w, http.StatusCreated, responses.User(r, u))
	}
}

const operationGetMyUser = "get_my_user"

func (c *UserController) GetMyUser(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMyUser)

	res, err := c.users.GetMyUserByID(r.Context(), scope.UserActor(r))
	switch {
	case errors.Is(err, errx.ErrorUserDeleted):
		log.WithError(err).Warn("user deleted")
		render.ResponseError(w, problems.Unauthorized())
		return
	case errors.Is(err, errx.ErrorUserNotFound):
		log.WithError(err).Warn("user not found")
		render.ResponseError(w, problems.NotFound("user not found"))
		return
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
		return
	}

	opts := make([]responses.UserOption, 0, 1)
	includes := restkit.ParseIncludes(r)

	if slices.Contains(includes, "email") {
		email, err := c.users.GetMyEmailByID(r.Context(), scope.UserActor(r))
		switch {
		case errors.Is(err, errx.ErrorUserDeleted):
			log.WithError(err).Warn("user deleted")
			render.ResponseError(w, problems.Unauthorized())
			return
		case errors.Is(err, errx.ErrorUserNotFound):
			log.WithError(err).Warn("user not found")
			render.ResponseError(w, problems.NotFound("user not found"))
			return
		case err != nil:
			log.WithError(err).Error("unexpected error")
			render.ResponseError(w, problems.InternalError())
			return
		}

		opts = append(opts, responses.WithUserEmail(email))
	}

	log.Info("user retrieved")
	render.Response(w, http.StatusOK, responses.User(r, res, opts...))
}

const operationGetUserByID = "get_user_by_id"

func (c *UserController) GetUserByID(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetUserByID)

	userID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		log.WithError(err).Warn("invalid user id")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"path": fmt.Errorf("invalid user id: %s", chi.URLParam(r, "user_id")),
		})...)
		return
	}

	log = log.With("target_user_id", userID)

	res, err := c.users.GetUserByID(r.Context(), userID)
	switch {
	case errors.Is(err, errx.ErrorUserNotFound):
		log.WithError(err).Warn("user does not exist")
		render.ResponseError(w, problems.NotFound("user does not exist"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.User(r, res))
	}
}

const operationGetUserByUsername = "get_user_by_username"

func (c *UserController) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetUserByUsername)

	username := chi.URLParam(r, "username")
	log = log.With("username", username)

	res, err := c.users.GetByUsername(r.Context(), username)
	switch {
	case errors.Is(err, errx.ErrorUserNotFound):
		log.WithError(err).Warn("user does not exist")
		render.ResponseError(w, problems.NotFound("user does not exist"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.User(r, res))
	}
}

const operationFilterUsers = "filter_users"

func (c *UserController) FilterUsers(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationFilterUsers)

	limit, offset := pagi.GetPagination(r)

	filters := user.FilterParams{}

	q := r.URL.Query()
	if text := strings.TrimSpace(q.Get("text")); text != "" {
		filters.Text = &text
	}

	res, err := c.users.GetList(r.Context(), filters, limit, offset)
	switch {
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.UserCollection(r, res))
	}
}

const operationUpdateMyUser = "update_my_user"

func (c *UserController) UpdateMyUser(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdateMyUser)

	req, err := requests.UpdateUser(r)
	if err != nil {
		log.WithError(err).Warn("invalid request body")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	res, err := c.users.Update(r.Context(), scope.UserActor(r), user.UpdateParams{
		AvatarKey:   req.Data.Attributes.AvatarKey,
		Pseudonym:   req.Data.Attributes.Pseudonym,
		Description: req.Data.Attributes.Description,
	})
	switch {
	case errors.Is(err, errx.ErrorUserNotFound):
		log.WithError(err).Warn("user not found")
		render.ResponseError(w, problems.NotFound("user not found"))
	case errors.Is(err, errx.ErrorUserUploadedAvatarInvalid):
		log.WithError(err).Warn("invalid avatar content")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"avatar": err,
		})...)
	case err != nil:
		log.WithError(err).Error("failed to update user")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Debug("user updated")
		render.Response(w, http.StatusOK, responses.User(r, res))
	}
}

const operationUpdateMyUsername = "update_my_username"

func (c *UserController) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdateMyUsername)

	req, err := requests.UpdateUsername(r)
	if err != nil {
		log.WithError(err).Warn("invalid request body")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	res, err := c.users.UpdateUsername(r.Context(), scope.UserActor(r), req.Data.Attributes.Username)
	switch {
	case errors.Is(err, errx.ErrorUserNotFound):
		log.WithError(err).Warn("user not found")
		render.ResponseError(w, problems.NotFound("user not found"))
	case errors.Is(err, errx.ErrorUsernameNotValid):
		log.WithError(err).Warn("username is invalid")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"username": err,
		})...)
	case errors.Is(err, errx.ErrorUsernameTaken):
		log.WithError(err).Warn("username is already taken")
		render.ResponseError(w, problems.Conflict("username is already taken"))
	case err != nil:
		log.WithError(err).Error("failed to update username")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Debug("username updated")
		render.Response(w, http.StatusOK, responses.User(r, res))
	}
}

const operationUpdatePassword = "update_password"

func (c *UserController) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdatePassword)

	req, err := requests.UpdatePassword(r)
	if err != nil {
		log.WithError(err).Info("failed to parse update password request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	err = c.users.UpdatePassword(
		r.Context(),
		scope.UserActor(r),
		req.Data.Attributes.OldPassword,
		req.Data.Attributes.NewPassword,
	)
	switch {
	case errors.Is(err, errx.ErrorUserInvalidSession),
		errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.WithError(err).Warn("invalid old password")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.WithError(err).Warn("new password is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/new_password": err,
		})...)
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("password updated successfully")
		render.Response(w, http.StatusNoContent, nil)
	}
}

const operationDeleteMyUser = "delete_my_user"

func (c *UserController) DeleteMyUser(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMyUser)

	err := c.users.DeleteMyUser(r.Context(), scope.UserActor(r))
	switch {
	case errors.Is(err, errx.ErrorUserInvalidSession),
		errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorUserNotFound):
		log.WithError(err).Warn("user not found")
		render.ResponseError(w, problems.NotFound("user not found"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("user deleted")
		render.Response(w, http.StatusNoContent, nil)
	}
}
