package outbound

import (
	"github.com/netbill/eventbox"
)

type Outbound struct {
	outbox eventbox.Producer
}

func New(producer eventbox.Producer) *Outbound {
	return &Outbound{
		outbox: producer,
	}
}
