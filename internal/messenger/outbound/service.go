package outbound

import (
	"github.com/netbill/auth-svc/internal/messenger"
	"github.com/netbill/eventbox"
)

type Outbound struct {
	groupID string
	outbox  eventbox.Producer
}

func New(producer eventbox.Producer) *Outbound {
	return &Outbound{
		groupID: messenger.AuthSvcGroup,
		outbox:  producer,
	}
}
