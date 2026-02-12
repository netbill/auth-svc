package messenger

import (
	"context"
	"fmt"
	"os"

	"github.com/netbill/auth-svc/internal/messenger/contracts"
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
	log *logium.Logger,
	db *pgdbx.DB,
	handlers handlers,
	config eventpg.InboxWorkerConfig,
) *Inbox {
	return &Inbox{
		log:      log.WithField("component", "inbox"),
		db:       db,
		handlers: handlers,
		config:   config,
	}
}

func (a *Inbox) Start(ctx context.Context) {
	a.log.Infoln("starting inbox worker")

	id := BuildProcessID("auth-svc", "inbox", 0)
	worker := eventpg.NewInboxWorker(a.log, a.db, id, a.config)

	defer func() {
		if err := worker.Stop(context.Background()); err != nil {
			a.log.WithError(err).Errorf("stop inbox worker %s failed", id)
		}
	}()

	worker.Route(contracts.OrgMemberCreatedEvent, a.handlers.OrgMemberCreated)
	worker.Route(contracts.OrgMemberDeletedEvent, a.handlers.OrgMemberDeleted)

	worker.Run(ctx)
}

func BuildProcessID(service string, role string, index int) string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return fmt.Sprintf("%s-%s-%d-%s", service, role, index, hostname)
}
