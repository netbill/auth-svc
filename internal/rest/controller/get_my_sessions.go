package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/pagi"
	"github.com/netbill/restkit/problems"
)

const operationGetMySessions = "get_my_sessions"

func (c *Controller) GetMySessions(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMySessions)

	limit, offset := pagi.GetPagination(r)

	sessions, err := c.core.GetMySessions(r.Context(), scope.AccountActor(r), limit, offset)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound) || errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Info("account not found by credentials")
		c.responser.RenderErr(w, problems.Unauthorized("account not found by credentials"))
	case err != nil:
		log.WithError(err).Error("failed to get my sessions")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		c.responser.Render(w, http.StatusOK, responses.AccountSessionsCollection(r, sessions))
	}
}
