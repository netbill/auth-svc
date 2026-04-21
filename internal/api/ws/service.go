package ws

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/netbill/auth-svc/pkg/log"
)

type core interface {
	CreateQRToken(ctx context.Context) (string, error)
}

type bus interface {
	SubscribeQRToken(ctx context.Context, key string) (<-chan []byte, func())
}

type Service struct {
	core core
	bus  bus

	upgrader websocket.Upgrader
	log      *log.Logger
}

func NewService(core core, messenger bus, log *log.Logger) *Service {
	return &Service{
		core: core,
		bus:  messenger,

		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		log: log,
	}
}
