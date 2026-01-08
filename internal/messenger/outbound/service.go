package outbound

import (
	"github.com/netbill/evebox/box/outbox"
	"github.com/netbill/logium"
)

type Producer struct {
	log    logium.Logger
	addr   []string
	outbox outbox.Box
}

func New(log logium.Logger, ob outbox.Box, addr ...string) *Producer {
	return &Producer{
		log:    log,
		addr:   addr,
		outbox: ob,
	}
}
