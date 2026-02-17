package messenger

import (
	"context"

	"github.com/netbill/auth-svc/pkg/evtypes"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/segmentio/kafka-go"
)

type handlers interface {
	OrgMemberCreated(
		ctx context.Context,
		message kafka.Message,
	) error
	OrgMemberDeleted(
		ctx context.Context,
		message kafka.Message,
	) error
}

func (m *Manager) RunInbox(ctx context.Context, handlers handlers) {
	id := BuildProcessID("inbox")
	worker := eventpg.NewInboxWorker(id, m.log, m.db, eventpg.InboxWorkerConfig{
		Routines:       m.config.Inbox.Routines,
		Slots:          m.config.Inbox.Slots,
		BatchSize:      m.config.Inbox.BatchSize,
		Sleep:          m.config.Inbox.Sleep,
		MinNextAttempt: m.config.Inbox.MinNextAttempt,
		MaxNextAttempt: m.config.Inbox.MaxNextAttempt,
		MaxAttempts:    m.config.Inbox.MaxAttempts,
	})

	defer func() {
		if err := worker.Stop(context.Background()); err != nil {
			m.log.WithError(err).Errorf("stop inbox worker %s failed", id)
		}
	}()

	worker.Route(evtypes.OrgMemberCreatedEvent, handlers.OrgMemberCreated)
	worker.Route(evtypes.OrgMemberDeletedEvent, handlers.OrgMemberDeleted)

	worker.Run(ctx)
}
