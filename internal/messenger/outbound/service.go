package outbound

import (
	"github.com/netbill/evebox/box/outbox"
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
)

type Outbound struct {
	log    *logium.Logger
	outbox outbox.Box
}

func New(log *logium.Logger, pool *pgdbx.DB) *Outbound {
	return &Outbound{
		log:    log,
		outbox: outbox.New(pool),
	}
}
