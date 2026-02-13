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
		log.WithError(err).Infof("invalid session id: %s", chi.URLParam(r, "session_id"))
		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"query": fmt.Errorf("invalid session id: %s", chi.URLParam(r, "session_id")),
		})...)

		return
	}

	log = log.WithField("target_session_id", sessionID)

	err = c.core.DeleteOwnSession(r.Context(), scope.AccountActor(r), sessionID)
	switch {
	case errors.Is(err, errx.ErrorInitiatorNotFound):
		log.Infof("initiator account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
	case errors.Is(err, errx.ErrorInitiatorInvalidSession):
		log.Infof("initiator session is invalid")
		c.responser.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
	case err != nil:
		log.WithError(err).Errorf("failed to delete My session")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		log.Info("initiator session deleted")
		c.responser.Render(w, http.StatusNoContent)
	}
}
