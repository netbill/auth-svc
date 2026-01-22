package messenger

import (
	"context"
	"sync"
	"time"

	"github.com/netbill/auth-svc/internal/messenger/contracts"
	"github.com/netbill/evebox/box/inbox"
	"github.com/netbill/evebox/consumer"
)

type handlers interface {
	OrgMemberCreated(
		ctx context.Context,
		event inbox.Event,
	) inbox.EventStatus
	OrgMemberDeleted(
		ctx context.Context,
		event inbox.Event,
	) inbox.EventStatus
}

func (m Messenger) RunConsumer(ctx context.Context, handlers handlers) {
	wg := &sync.WaitGroup{}
	run := func(f func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f()
		}()
	}

	orgConsumer := consumer.New(m.log, m.db, "auth-svc-org-consumer", consumer.OnUnknownDoNothing, m.addr...)

	orgConsumer.Handle(contracts.OrgMemberCreatedEvent, handlers.OrgMemberCreated)
	orgConsumer.Handle(contracts.OrgMemberDeletedEvent, handlers.OrgMemberDeleted)

	inboxer1 := consumer.NewInboxer(m.log, m.db, consumer.ConfigInboxer{
		Name:       "auth-svc-inbox-worker-1",
		BatchSize:  10,
		RetryDelay: 1 * time.Minute,
		MinSleep:   100 * time.Millisecond,
		MaxSleep:   1 * time.Second,
	})
	inboxer1.Handle(contracts.OrgMemberCreatedEvent, handlers.OrgMemberCreated)
	inboxer1.Handle(contracts.OrgMemberDeletedEvent, handlers.OrgMemberDeleted)

	run(func() {
		orgConsumer.Run(ctx, contracts.AuthSvcGroup, contracts.AccountsTopicV1, m.addr...)
	})

	run(func() {
		inboxer1.Run(ctx)
	})

	wg.Wait()
}
