package controller

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/contexter"
	"github.com/netbill/restkit/problems"
)

func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	initiator, err := contexter.AccountData(r.Context())
	if err != nil {
		c.log.WithError(err).Error("failed to get user from context")
		c.responser.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	err = c.core.Logout(r.Context(), account.InitiatorData{
		AccountID: initiator.GetAccountID(),
		SessionID: initiator.GetSessionID(),
	})
	if err != nil {
		c.log.WithError(err).Errorf("failed to logout user")
		switch {
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.responser.Render(w, http.StatusNoContent)
}
