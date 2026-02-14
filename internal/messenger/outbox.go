package messenger

import (
	"context"

	eventpg "github.com/netbill/eventbox/pg"
	"github.com/segmentio/kafka-go"
)

func (m *Manager) RunOutbox(ctx context.Context) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(m.config.Brokers...),
		RequiredAcks: kafka.RequireAll,
		Compression:  kafka.Snappy,
		Balancer:     &kafka.LeastBytes{},
	}

	defer func() {
		if err := writer.Close(); err != nil {
			m.log.WithError(err).Error("failed to close kafka writer")
		}
	}()

	id := BuildProcessID("outbox")
	worker := eventpg.NewOutboxWorker(id, m.log, m.db, writer, eventpg.OutboxWorkerConfig{
		Routines:       m.config.Outbox.Routines,
		Slots:          m.config.Outbox.Slots,
		BatchSize:      m.config.Outbox.BatchSize,
		Sleep:          m.config.Outbox.Sleep,
		MinNextAttempt: m.config.Outbox.MinNextAttempt,
		MaxNextAttempt: m.config.Outbox.MaxNextAttempt,
		MaxAttempts:    m.config.Outbox.MaxAttempts,
	})

	defer func() {
		if err := worker.Stop(context.Background()); err != nil {
			m.log.WithError(err).Errorf("stop outbox worker %s failed", id)
		}
	}()

	worker.Run(ctx)
}
