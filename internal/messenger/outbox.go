package messenger

import (
	"context"

	eventpg "github.com/netbill/eventbox/pg"
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
	"github.com/segmentio/kafka-go"
)

type Outbox struct {
	log    *logium.Entry
	db     *pgdbx.DB
	addr   []string
	config eventpg.OutboxWorkerConfig
}

func NewOutbox(
	log *logium.Logger,
	db *pgdbx.DB,
	addr []string,
	config eventpg.OutboxWorkerConfig,
) *Outbox {
	return &Outbox{
		db:     db,
		log:    log.WithField("component", "outbox"),
		addr:   addr,
		config: config,
	}
}

func (a *Outbox) Start(ctx context.Context) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(a.addr...),
		RequiredAcks: kafka.RequireAll,
		Compression:  kafka.Snappy,
		Balancer:     &kafka.LeastBytes{},
	}
	defer func() {
		if err := writer.Close(); err != nil {
			a.log.WithError(err).Error("failed to close kafka writer")
		}
	}()

	a.log.Infoln("starting outbox worker")

	id := BuildProcessID("auth-svc", "outbox", 0)
	worker := eventpg.NewOutboxWorker(a.log, a.db, writer, id, a.config)

	defer func() {
		if err := worker.Stop(context.Background()); err != nil {
			a.log.WithError(err).WithField("worker_id", id).Error("failed to stop outbox worker")
		}
	}()

	worker.Run(ctx)
}
