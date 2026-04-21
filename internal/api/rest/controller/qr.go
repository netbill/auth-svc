package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/api/rest/requests"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

type handler interface {
	LoginWithQR(w http.ResponseWriter, r *http.Request, h http.Header)
}

type sessions interface {
	ConfirmQRToken(
		ctx context.Context,
		actor models.AccountActor,
		qrToken string,
	) (models.TokensPair, error)
}

type bus interface {
	PublishQRToken(ctx context.Context, key string, payload []byte) error
}

type QRController struct {
	wsHandler handler
	sessions  sessions
	bus       bus
}

func NewQRController(hand handler, sessions sessions, bus bus) *QRController {
	return &QRController{
		wsHandler: hand,
		sessions:  sessions,
		bus:       bus,
	}
}

func (c *QRController) QRConnect(w http.ResponseWriter, r *http.Request) {
	c.wsHandler.LoginWithQR(w, r, nil)
}

const operationQRConfirm = "qr_confirm"

func (c *QRController) QRConfirm(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationQRConfirm)

	req, err := requests.QRConfirm(r)
	if err != nil {
		log.WithError(err).Warn("invalid qr confirm request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	qrToken := req.Data.Attributes.QrToken.String()

	pair, err := c.sessions.ConfirmQRToken(r.Context(), scope.AccountActor(r), qrToken)
	switch {
	case errors.Is(err, errx.ErrorQRTokenNotFound):
		log.WithError(err).Warn("qr token not found")
		render.ResponseError(w, problems.NotFound("qr token not found or expired"))
		return
	case errors.Is(err, errx.ErrorQRTokenAlreadyConfirmed):
		log.WithError(err).Warn("qr token already confirmed")
		render.ResponseError(w, problems.Conflict("qr token already confirmed"))
		return
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
		return
	}

	payload, err := json.Marshal(pair)
	if err != nil {
		log.WithError(err).Error("failed to marshal tokens pair")
		render.ResponseError(w, problems.InternalError())
		return
	}

	if err = c.bus.PublishQRToken(r.Context(), qrToken, payload); err != nil {
		log.WithError(err).Error("failed to publish qr confirm")
		render.ResponseError(w, problems.InternalError())
		return
	}

	log.Info("qr token confirmed")
	render.Response(w, http.StatusNoContent, nil)
}
