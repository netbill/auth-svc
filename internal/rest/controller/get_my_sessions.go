package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/contexter"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/pagi"
	"github.com/netbill/restkit/problems"
)

func (c *Controller) GetMySessions(w http.ResponseWriter, r *http.Request) {
	initiator, err := contexter.AccountData(r.Context())
	if err != nil {
		c.log.WithError(err).Error("failed to get user from context")
		c.responser.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	limit, offset := pagi.GetPagination(r)
	sessions, err := c.core.GetOwnSessions(r.Context(), account.InitiatorData{
		AccountID: initiator.GetAccountID(),
		SessionID: initiator.GetSessionID(),
	}, limit, offset)
	if err != nil {
		c.log.WithError(err).Errorf("failed to select My sessions")
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			c.responser.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.responser.Render(w, http.StatusOK, responses.AccountSessionsCollection(r, sessions))
}
