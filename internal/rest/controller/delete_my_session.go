package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
)

const operationDeleteMySession = "delete_my_session"

func (c *Controller) DeleteMySession(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMySession)

	sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
	if err != nil {
		log.WithError(err).WithField("session_id", chi.URLParam(r, "session_id")).Error("invalid session id")
		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"query": fmt.Errorf("invalid session id: %s", chi.URLParam(r, "session_id")),
		})...)

		return
	}

	log = log.WithField("target_session_id", sessionID)

	err = c.core.DeleteMySession(r.Context(), scope.AccountActor(r), sessionID)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound) || errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Info("account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("account not found by credentials"))
	case err != nil:
		log.WithError(err).Error("failed to delete My session")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		log.Info("session deleted")
		c.responser.Status(w, http.StatusNoContent)
	}
}
