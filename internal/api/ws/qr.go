package ws

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/netbill/auth-svc/internal/api/ws/responses"
)

// LoginWithQR содержит всю логику WS сессии QR-логина.
// Контроллер только апгрейдит соединение и вызывает этот метод.
func (s *Service) LoginWithQR(w http.ResponseWriter, r *http.Request, responseHeader http.Header) {
	conn, err := s.upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		s.log.WithError(err).Warn("ws upgrade failed")
		return
	}
	defer conn.Close()

	token, err := s.core.CreateQRToken(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to create qr token")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	msgCh, cleanup := s.bus.SubscribeQRToken(ctx, token)
	defer cleanup()

	if err = responses.QRToken(conn, token); err != nil {
		s.log.WithError(err).Error("failed to send QR token")
		return
	}

	go func() {
		for {
			if _, _, err = conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	select {
	case payload := <-msgCh:
		if err = responses.AuthTokensPair(conn, payload); err != nil {
			s.log.WithError(err).Error("failed to write tokens pair")
		}
	case <-ctx.Done():
		if err = responses.TokensExpiredError(conn); err != nil {
			s.log.WithError(err).Error("failed to write tokens expired error")
		}
	}

	if err = conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	); err != nil {
		s.log.WithError(err).Error("failed to write close message")
	}
}
