package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

const operationGetMySession = "get_my_session"

func (c *Controller) GetMySession(w http.ResponseWriter, r *http.Request) {
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

	session, err := c.core.GetMySession(r.Context(), scope.AccountActor(r), sessionID)
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Info("invalid credentials")
		render.ResponseError(w, problems.Unauthorized("invalid credentials"))
	case errors.Is(err, errx.ErrorSessionDeleted) || errors.Is(err, errx.ErrorSessionNotFound):
		log.Info("session not found")
		render.ResponseError(w, problems.NotFound("session not found"))
	case err != nil:
		log.WithError(err).Error("failed to get my session")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.AccountSession(session))
	}
}
