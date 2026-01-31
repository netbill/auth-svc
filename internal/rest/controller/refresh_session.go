package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/restkit/problems"
)

func (c *Controller) RefreshSession(w http.ResponseWriter, r *http.Request) {
	req, err := requests.RefreshSession(r)
	if err != nil {
		c.log.WithError(err).Error("failed to parse refresh session request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)

		return
	}

	tokensPair, err := c.core.Refresh(r.Context(), req.Data.Attributes.RefreshToken)
	if err != nil {
		c.log.WithError(err).Errorf("failed to refresh session token")
		switch {
		case errors.Is(err, errx.ErrorAccountNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("account not found"))
		case errors.Is(err, errx.ErrorSessionNotFound):
			c.responser.RenderErr(w, problems.Unauthorized("session not found"))
		case errors.Is(err, errx.ErrorSessionTokenMismatch):
			c.responser.RenderErr(w, problems.Forbidden("refresh session token mismatch"))
		default:
			c.responser.RenderErr(w, problems.InternalError())
		}

		return
	}

	c.responser.Render(w, http.StatusOK, responses.TokensPair(tokensPair))
}
