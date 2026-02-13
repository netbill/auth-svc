package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
)

const operationGetMyEmailData = "get_my_email_data"

func (c *Controller) GetMyEmailData(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMyEmailData)

	emailData, err := c.core.GetAccountEmail(r.Context(), scope.AccountActor(r).ID)
	switch {
	case errors.Is(err, errx.ErrorAccountEmailNotFound):
		log.Info("initiator account email not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("initiator account not found by credentials"))
	case err != nil:
		log.WithError(err).Error("failed to get my email data")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		c.responser.Render(w, http.StatusOK, responses.AccountEmailData(emailData))
	}
}
