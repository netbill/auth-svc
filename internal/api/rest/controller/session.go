package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/api/rest/requests"
	"github.com/netbill/auth-svc/internal/api/rest/responses"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/restkit/pagi"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
	"golang.org/x/oauth2"
)

type sessionCore interface {
	LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error)
	LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error)
	LoginByUsername(ctx context.Context, username, password string) (models.TokensPair, error)

	Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error)

	GetMySession(ctx context.Context, actor models.AccountActor, sessionID uuid.UUID) (models.Session, error)
	GetMySessions(
		ctx context.Context,
		actor models.AccountActor,
		opts ...session.ListSessionsOption,
	) (pagi.Page[[]models.Session], error)

	ConfirmQRToken(
		ctx context.Context,
		actor models.AccountActor,
		qrToken string,
	) (models.TokensPair, error)

	PublishQRToken(ctx context.Context, key string, payload []byte) error

	Logout(ctx context.Context, actor models.AccountActor) error
	DeleteMySession(ctx context.Context, actor models.AccountActor, sessionID uuid.UUID) error
	DeleteMySessions(ctx context.Context, actor models.AccountActor) error
}

type SessionMetrics interface {
	RecordEmailLogin(ctx context.Context, err *error)
	RecordUsernameLogin(ctx context.Context, err *error)
	RecordGoogleLogin(ctx context.Context, err *error)
	RecordTokenRefresh(ctx context.Context, err *error)
	RecordSessionDeleted(ctx context.Context, scope string, err *error)
	RecordQRLogin(ctx context.Context, err *error)
}

type handler interface {
	LoginWithQR(w http.ResponseWriter, r *http.Request, h http.Header)
}

type SessionController struct {
	google    oauth2.Config
	sessions  sessionCore
	metrics   SessionMetrics
	wsHandler handler
}

func NewSessionController(
	sessions sessionCore,
	google oauth2.Config,
	m SessionMetrics,
	wsHandler handler,
) *SessionController {
	return &SessionController{
		google:    google,
		sessions:  sessions,
		metrics:   m,
		wsHandler: wsHandler,
	}
}

const operationGetMySession = "get_my_session"

func (c *SessionController) GetMySession(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMySession)

	sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
	if err != nil {
		log.WithError(err).
			WithField("target_session_id", chi.URLParam(r, "session_id")).
			Warn("invalid session id")

		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"path": fmt.Errorf("invalid session id: %s", chi.URLParam(r, "session_id")),
		})...)
		return
	}

	log = log.WithField("target_session_id", sessionID)

	s, err := c.sessions.GetMySession(r.Context(), scope.AccountActor(r), sessionID)
	switch {
	case errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("session not found")
		render.ResponseError(w, problems.NotFound("session not found"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("session retrieved")
		render.Response(w, http.StatusOK, responses.AccountSession(s))
	}
}

const operationGetMySessions = "get_my_sessions"

func (c *SessionController) GetMySessions(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMySessions)

	limit, offset := pagi.GetPagination(r)

	opts := []session.ListSessionsOption{
		session.WithLimit(limit),
		session.WithOffset(offset),
	}

	switch r.URL.Query().Get("filter[active]") {
	case "true":
		opts = append(opts, session.WithDeleted(session.DeletedFilterActive))
	case "false":
		opts = append(opts, session.WithDeleted(session.DeletedFilterDeleted))
	default:
		opts = append(opts, session.WithDeleted(session.DeletedFilterAll))
	}

	switch r.URL.Query().Get("sort[last_used]") {
	case "asc":
		opts = append(opts, session.WithLastUsedOrder(session.LastUsedAsc))
	default:
		opts = append(opts, session.WithLastUsedOrder(session.LastUsedDesc))
	}

	sessions, err := c.sessions.GetMySessions(r.Context(), scope.AccountActor(r), opts...)
	switch {
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("sessions retrieved")
		render.Response(w, http.StatusOK, responses.AccountSessionsCollection(r, sessions))
	}
}

const operationRefreshSession = "refresh_session"

func (c *SessionController) RefreshSession(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRefreshSession)

	req, err := requests.RefreshSession(r)
	if err != nil {
		log.WithError(err).Info("invalid refresh session request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	defer c.metrics.RecordTokenRefresh(r.Context(), &err)

	tokensPair, err := c.sessions.Refresh(r.Context(), req.Data.Attributes.RefreshToken)
	switch {
	case errors.Is(err, errx.ErrorSessionExpired):
		log.WithError(err).Warn("refresh token expired")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.WithError(err).Warn("account not found")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("session not found")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorSessionTokenMismatch):
		log.WithError(err).Warn("refresh token mismatch")
		render.ResponseError(w, problems.Unauthorized())
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("session refreshed successfully")
		render.Response(w, http.StatusOK, responses.TokensPair(tokensPair))
	}
}

const operationUpdateUsername = "update_username"

func (c *AccountController) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdateUsername)

	req, err := requests.UpdateUsername(r)
	if err != nil {
		log.WithError(err).Info("failed to parse update username request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	log = log.WithField("new_username", req.Data.Attributes.Username)

	res, err := c.accounts.UpdateUsername(r.Context(), scope.AccountActor(r), req.Data.Attributes.Username)
	switch {
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.WithError(err).Info("username is not allowed")
		render.ResponseError(w, problems.BadRequest(err)...)
	case errors.Is(err, errx.ErrorAccountDeleted):
		log.WithError(err).Warn("account deleted")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.WithError(err).Warn("account not found")
		render.ResponseError(w, problems.NotFound("account not found"))
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.WithError(err).Warn("username is already taken")
		render.ResponseError(w, problems.Conflict("user with this username already exists"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("username updated successfully")
		render.Response(w, http.StatusOK, responses.Account(res))
	}
}

const operationDeleteMySession = "delete_my_session"

func (c *SessionController) DeleteMySession(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMySession)

	sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
	if err != nil {
		log.WithError(err).
			WithField("session_id", chi.URLParam(r, "session_id")).
			Error("invalid session id")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"query": fmt.Errorf("invalid session id: %s", chi.URLParam(r, "session_id")),
		})...)

		return
	}

	log = log.WithField("target_session_id", sessionID)

	defer c.metrics.RecordSessionDeleted(r.Context(), "single", &err)

	err = c.sessions.DeleteMySession(r.Context(), scope.AccountActor(r), sessionID)
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("session not found")
		render.ResponseError(w, problems.NotFound("session not found"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("session deleted")
		render.Response(w, http.StatusNoContent, nil)
	}
}

const operationDeleteMySessions = "delete_my_sessions"

func (c *SessionController) DeleteMySessions(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMySessions)

	var err error
	defer c.metrics.RecordSessionDeleted(r.Context(), "all", &err)

	err = c.sessions.DeleteMySessions(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession),
		errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("sessions deleted")
		render.Response(w, http.StatusNoContent, nil)
	}
}

const operationLogout = "logout"

func (c *SessionController) Logout(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationLogout)

	err := c.sessions.Logout(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorSessionNotFound):
		// already logged out — treat as success
		log.WithError(err).Debug("session already deleted, treating logout as success")
		render.Response(w, http.StatusNoContent, nil)
	case err != nil:
		log.WithError(err).Error("logout failed")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("logout successful")
		render.Response(w, http.StatusNoContent, nil)
	}
}
