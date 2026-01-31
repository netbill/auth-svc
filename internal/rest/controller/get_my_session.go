package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/contexter"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/problems"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

func (c *Controller) GetMySession(w http.ResponseWriter, r *http.Request) {
	initiator, err := contexter.AccountData(r.Context())
	if err != nil {
		c.log.WithError(err).Error("failed to get user from context")
		c.responser.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	sessionId, err := uuid.Parse(chi.URLParam(r, "session_id"))
	if err != nil {
		c.log.WithError(err).Errorf("invalid session id: %s", chi.URLParam(r, "session_id"))
		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"query": fmt.Errorf("invalid session id: %s", chi.URLParam(r, "session_id")),
		})...)

		return
	}

	session, err := c.core.GetOwnSession(r.Context(), account.InitiatorData{
		AccountID: initiator.GetAccountID(),
		SessionID: initiator.GetSessionID(),
	}, sessionId)
	if err != nil {
		c.log.WithError(err).Errorf("failed to get My session")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		case errors.Is(err, errx.ErrorSessionNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("session not found"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			c.responser.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.responser.Render(w, http.StatusOK, responses.AccountSession(session))
}
