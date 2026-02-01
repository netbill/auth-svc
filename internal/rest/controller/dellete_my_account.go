package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/contexter"
	"github.com/netbill/restkit/problems"
)

func (c *Controller) DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	initiator, err := contexter.AccountData(r.Context())
	if err != nil {
		c.log.WithError(err).Error("failed to get user from context")
		c.responser.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	err = c.core.DeleteOwnAccount(r.Context(), account.InitiatorData{
		AccountID: initiator.GetAccountID(),
		SessionID: initiator.GetSessionID(),
	})
	if err != nil {
		c.log.WithError(err).Errorf("failed to delete my account with id: %s", initiator.GetAccountID())
		switch {
		case errors.Is(err, errx.ErrorInitiatorNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		case errors.Is(err, errx.ErrorInitiatorInvalidSession):
			c.responser.RenderErr(w, problems.Unauthorized("initiator session is invalid"))
		case errors.Is(err, errx.AccountHaveMembershipInOrg):
			c.responser.RenderErr(w, problems.Forbidden("account cannot be deleted while having membership in organization"))
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.responser.Render(w, http.StatusNoContent)
}
