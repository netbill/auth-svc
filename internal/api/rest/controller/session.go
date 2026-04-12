package controller

import (
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
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/restkit/pagi"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

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

	err := c.sessions.DeleteMySessions(r.Context(), scope.AccountActor(r))
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
		render.Response(w, http.StatusNoContent, nil)
	}
}
