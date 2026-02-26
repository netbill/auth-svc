package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationRefreshSession = "refresh_session"

func (c *Controller) RefreshSession(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRefreshSession)

	req, err := requests.RefreshSession(r)
	if err != nil {
		log.WithError(err).Info("invalid refresh session request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	tokensPair, err := c.core.Refresh(r.Context(), req.Data.Attributes.RefreshToken)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.Info("account not found")
		render.ResponseError(w, problems.Unauthorized("account not found"))
	case errors.Is(err, errx.ErrorSessionNotFound):
		log.Info("session not found")
		render.ResponseError(w, problems.Unauthorized("session not found"))
	case errors.Is(err, errx.ErrorSessionTokenMismatch):
		log.Info("refresh session token mismatch")
		render.ResponseError(w, problems.Forbidden("refresh session token mismatch"))
	case err != nil:
		log.WithError(err).Error("refresh session failed")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.TokensPair(tokensPair))
	}
}
