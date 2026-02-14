package messenger

import (
	"context"
	"sync"

	"github.com/netbill/auth-svc/internal/messenger/evtypes"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/segmentio/kafka-go"
)

func (m *Manager) RunConsumer(ctx context.Context) {
	var wg sync.WaitGroup

	consumer := eventpg.NewConsumer(m.log, m.db, eventpg.ConsumerConfig{})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        m.config.Brokers,
		Topic:          evtypes.OrgMemberTopicV1,
		GroupID:        evtypes.AuthSvcGroup,
		QueueCapacity:  m.config.Reader.Topics.OrgMembersV1.QueueCapacity,
		MaxBytes:       m.config.Reader.Topics.OrgMembersV1.MaxBytes,
		MinBytes:       m.config.Reader.Topics.OrgMembersV1.MinBytes,
		MaxWait:        m.config.Reader.Topics.OrgMembersV1.MaxWait,
		CommitInterval: m.config.Reader.Topics.OrgMembersV1.CommitInterval,
	})

	wg.Add(1)
	go func(r *kafka.Reader) {
		defer r.Close()
		defer wg.Done()

		consumer.Read(ctx, r)
	}(reader)

	wg.Wait()
}
