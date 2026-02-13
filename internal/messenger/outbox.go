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
	log *logium.Entry,
	db *pgdbx.DB,
	addr []string,
	config eventpg.OutboxWorkerConfig,
) *Outbox {
	return &Outbox{
		db:     db,
		log:    log.WithComponent("outbox"),
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

	id := BuildProcessID("outbox")
	worker := eventpg.NewOutboxWorker(id, a.log, a.db, writer, a.config)

	defer func() {
		if err := worker.Stop(context.Background()); err != nil {
			a.log.WithError(err).WithField("worker_id", id).Error("failed to stop outbox worker")
		}
	}()

	worker.Run(ctx)
}
