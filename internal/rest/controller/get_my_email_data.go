package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationGetMyEmailData = "get_my_email_data"

func (c *Controller) GetMyEmailData(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMyEmailData)

	emailData, err := c.core.GetMyAccountEmail(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Info("invalid credentials")
		render.ResponseError(w, problems.Unauthorized("invalid credentials"))
	case err != nil:
		log.WithError(err).Error("failed to get my email data")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.AccountEmailData(emailData))
	}
}
