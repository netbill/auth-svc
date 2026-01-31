package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/contexter"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/problems"
)

func (c *Controller) GetMyEmailData(w http.ResponseWriter, r *http.Request) {
	initiator, err := contexter.AccountData(r.Context())
	if err != nil {
		c.log.WithError(err).Error("failed to get user from context")
		c.responser.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	emailData, err := c.core.GetAccountEmail(r.Context(), initiator.GetAccountID())
	if err != nil {
		c.log.WithError(err).Errorf("failed to get email repo by id: %s", initiator.GetAccountID())
		switch {
		case errors.Is(err, errx.ErrorAccountEmailNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.responser.Render(w, http.StatusOK, responses.AccountEmailData(emailData))
}
