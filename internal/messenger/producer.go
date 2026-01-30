package messenger

import (
	"context"
	"sync"
	"time"

	"github.com/netbill/evebox/producer"
	"github.com/segmentio/kafka-go"
)

func (m Messenger) RunProducer(ctx context.Context) {
	wg := &sync.WaitGroup{}

	run := func(f func()) {
		wg.Add(1)
		go func() {
			f()
			wg.Done()
		}()
	}

	worker1 := producer.New(producer.NewProducerParams{
		Name:            "outbox-worker-1",
		KafkaAddr:       m.addr,
		Log:             m.log,
		DB:              m.db,
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

	worker2 := producer.New(producer.NewProducerParams{
		Name:            "outbox-worker-2",
		KafkaAddr:       m.addr,
		Log:             m.log,
		DB:              m.db,
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
