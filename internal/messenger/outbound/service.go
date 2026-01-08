package outbound

import (
	"github.com/netbill/evebox/box/outbox"
	"github.com/netbill/logium"
)

type Producer struct {
	log    logium.Logger
	outbox outbox.Box
}

func New(log logium.Logger, ob outbox.Box) *Producer {
	return &Producer{
		log:    log,
		outbox: ob,
	}
}
