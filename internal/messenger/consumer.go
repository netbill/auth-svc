package messenger

import (
	"context"
	"sync"

	evtypes2 "github.com/netbill/auth-svc/pkg/evtypes"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/segmentio/kafka-go"
)

func (m *Manager) RunConsumer(ctx context.Context) {
	var wg sync.WaitGroup

	consumer := eventpg.NewConsumer(m.log, m.db, eventpg.ConsumerConfig{})

	orgMemberReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        m.config.Brokers,
		Topic:          evtypes2.OrgMemberTopicV1,
		GroupID:        AuthSvcGroup,
		QueueCapacity:  m.config.Reader.Topics.OrgMembersV1.QueueCapacity,
		MaxBytes:       m.config.Reader.Topics.OrgMembersV1.MaxBytes,
		MinBytes:       m.config.Reader.Topics.OrgMembersV1.MinBytes,
		MaxWait:        m.config.Reader.Topics.OrgMembersV1.MaxWait,
		CommitInterval: m.config.Reader.Topics.OrgMembersV1.CommitInterval,
	})

	orgReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        m.config.Brokers,
		Topic:          evtypes2.OrgTopicV1,
		GroupID:        AuthSvcGroup,
		QueueCapacity:  m.config.Reader.Topics.OrganizationsV1.QueueCapacity,
		MaxBytes:       m.config.Reader.Topics.OrganizationsV1.MaxBytes,
		MinBytes:       m.config.Reader.Topics.OrganizationsV1.MinBytes,
		MaxWait:        m.config.Reader.Topics.OrganizationsV1.MaxWait,
		CommitInterval: m.config.Reader.Topics.OrganizationsV1.CommitInterval,
	})

	run := func(r *kafka.Reader) {
		wg.Add(1)
		go func() {
			defer r.Close()
			defer wg.Done()

			consumer.Read(ctx, r)
		}()
	}

	run(orgMemberReader)
	run(orgReader)

	wg.Wait()
}
