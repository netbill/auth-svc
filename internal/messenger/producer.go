package messenger

import (
	"context"
	"sync"
	"time"

	"github.com/netbill/evebox/box/outbox"
	"github.com/netbill/evebox/producer"
	"github.com/netbill/logium"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	log    logium.Logger
	addr   []string
	outbox outbox.Box
}

func NewProducer(log logium.Logger, ob outbox.Box, addr ...string) *Producer {
	return &Producer{
		log:    log,
		addr:   addr,
		outbox: ob,
	}
}

func (p Producer) Run(ctx context.Context) {
	wg := &sync.WaitGroup{}

	run := func(f func()) {
		wg.Add(1)
		go func() {
			f()
			wg.Done()
		}()
	}

	worker1 := producer.NewOutboxWorker(p.log, p.outbox, p.addr, producer.OutboxWorkerConfig{
		Name:            "outbox-worker-1",
		BatchLimit:      10,
		LockTTL:         30 * time.Second,
		EventRetryDelay: 1 * time.Minute,
		MinSleep:        100 * time.Millisecond,
		MaxSleep:        1 * time.Second,
		RequiredAcks:    kafka.RequireAll,
		Compression:     kafka.Snappy,
		BatchTimeout:    50,
		Balancer:        &kafka.LeastBytes{},
	})

	worker2 := producer.NewOutboxWorker(p.log, p.outbox, p.addr, producer.OutboxWorkerConfig{
		Name:            "outbox-worker-2",
		BatchLimit:      10,
		LockTTL:         30 * time.Second,
		EventRetryDelay: 1 * time.Minute,
		MinSleep:        100 * time.Millisecond,
		MaxSleep:        1 * time.Second,
		RequiredAcks:    kafka.RequireAll,
		Compression:     kafka.Snappy,
		BatchTimeout:    50,
		Balancer:        &kafka.LeastBytes{},
	})

	run(func() { worker1.Run(ctx) })

	run(func() { worker2.Run(ctx) })

	wg.Wait()
}
