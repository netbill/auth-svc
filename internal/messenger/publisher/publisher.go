package publisher

import (
	"github.com/netbill/eventbox"
)

type Publisher struct {
	identity string
	outbox   eventbox.Outbox
	producer *eventbox.Producer
}

func New(
	identity string,
	outbox eventbox.Outbox,
	producer *eventbox.Producer,
) *Publisher {
	return &Publisher{
		identity: identity,
		outbox:   outbox,
		producer: producer,
	}
}
