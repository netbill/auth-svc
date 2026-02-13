package messenger

import (
	"context"
	"fmt"
	"os"

	"github.com/netbill/auth-svc/internal/messenger/evtypes"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
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

type Inbox struct {
	log      *logium.Entry
	db       *pgdbx.DB
	handlers handlers
	config   eventpg.InboxWorkerConfig
}

func NewInbox(
	log *logium.Entry,
	db *pgdbx.DB,
	handlers handlers,
	config eventpg.InboxWorkerConfig,
) *Inbox {
	return &Inbox{
		log:      log.WithComponent("inbox"),
		db:       db,
		handlers: handlers,
		config:   config,
	}
}

func (b *Inbox) Start(ctx context.Context) {
	b.log.Infoln("starting inbox worker")

	id := BuildProcessID("inbox")
	worker := eventpg.NewInboxWorker(id, b.log, b.db, b.config)

	defer func() {
		if err := worker.Stop(context.Background()); err != nil {
			b.log.WithError(err).Errorf("stop inbox worker %s failed", id)
		}
	}()

	worker.Route(evtypes.OrgMemberCreatedEvent, b.handlers.OrgMemberCreated)
	worker.Route(evtypes.OrgMemberDeletedEvent, b.handlers.OrgMemberDeleted)

	worker.Run(ctx)
}

func BuildProcessID(service string) string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return fmt.Sprintf("%s-%s-%d", service, hostname, os.Getpid())
}
