package messenger

import (
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/eventbox"
)

func NewOutboxWorker(
	log *log.Logger,
	outbox eventbox.Outbox,
	producer *eventbox.Producer,
	cfg eventbox.OutboxWorkerConfig,
) *eventbox.OutboxWorker {
	return eventbox.NewOutboxWorker(
		uuid.New().String(),
		log, outbox, producer, cfg,
	)
}
