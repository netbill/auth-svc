package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"

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

		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"path": fmt.Errorf("invalid session id: %s", chi.URLParam(r, "session_id")),
		})...)
		return
	}

	log = log.WithField("target_session_id", sessionID)

	session, err := c.core.GetMySession(r.Context(), scope.AccountActor(r), sessionID)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountInvalidSession),
		errors.Is(err, errx.ErrorSessionNotFound):

		log.Infof("account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("account not found by credentials"))
	case err != nil:
		log.WithError(err).Error("failed to get my session")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		c.responser.Render(w, http.StatusOK, responses.AccountSession(session))
	}
}
