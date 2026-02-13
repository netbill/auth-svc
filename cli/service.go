package cli

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/core/modules/organization"
	"github.com/netbill/auth-svc/internal/messenger"
	"github.com/netbill/auth-svc/internal/messenger/evtypes"
	"github.com/netbill/auth-svc/internal/messenger/inbound"
	"github.com/netbill/auth-svc/internal/messenger/outbound"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/auth-svc/internal/repository/pg"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/rest/controller"
	"github.com/netbill/auth-svc/internal/rest/middlewares"
	"github.com/netbill/auth-svc/internal/tokenmanger"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
	"github.com/netbill/restkit"
)

func StartServices(ctx context.Context, cfg *Config, log *logium.Entry, wg *sync.WaitGroup) {
	run := func(f func()) {
		wg.Add(1)
		go func() {
			f()
			wg.Done()
		}()
	}

	pool, err := pgxpool.New(ctx, cfg.Database.SQL.URL)
	if err != nil {
		log.Fatal("failed to connect to database", "error", err)
	}
	db := pgdbx.NewDB(pool)

	repo := &repository.Repository{
		AccountsQ:      pg.NewAccountsQ(db),
		AccountEmailsQ: pg.NewAccountEmailsQ(db),
		AccountPassQ:   pg.NewAccountPasswordsQ(db),
		SessionsQ:      pg.NewSessionsQ(db),
		OrgMembersQ:    pg.NewOrganizationMembersQ(db),
		Transactioner:  pg.NewTransaction(db),
	}

	kafkaProducer := messenger.NewProducer(db, cfg.Kafka.Brokers...)
	kafkaOutbound := outbound.New(kafkaProducer)

	tokenManager := &tokenmanger.Manager{
		Issuer:     ServiceName,
		AccessSK:   cfg.Auth.Account.Token.Access.SecretKey,
		RefreshSK:  cfg.Auth.Account.Token.Refresh.SecretKey,
		RefreshHK:  cfg.Auth.Account.Token.Refresh.HashKey,
		AccessTTL:  cfg.Auth.Account.Token.Access.Lifetime,
		RefreshTTL: cfg.Auth.Account.Token.Refresh.Lifetime,
	}

	authModule := account.NewService(repo, tokenManager, kafkaOutbound)
	orgModule := organization.New(repo)

	responser := restkit.NewResponser()
	ctrl := controller.New(authModule, cfg.GoogleOAuth(), responser)
	mdll := middlewares.New(tokenManager, responser)
	router := rest.New(mdll, ctrl)

	run(func() {
		router.Run(ctx, log, rest.Config{
			Port:              cfg.Rest.Port,
			TimeoutRead:       cfg.Rest.Timeouts.Read,
			TimeoutReadHeader: cfg.Rest.Timeouts.ReadHeader,
			TimeoutWrite:      cfg.Rest.Timeouts.Write,
			TimeoutIdle:       cfg.Rest.Timeouts.Idle,
		})
	})

	kafkaConsumer := messenger.NewConsumer(log, db, cfg.Kafka.Brokers, map[string]int{
		evtypes.OrgMemberTopicV1: cfg.Kafka.Readers.OrgMemberV1,
	})

	run(func() { kafkaConsumer.Start(ctx) })

	kafkaInboxArh := messenger.NewInbox(log, db, inbound.New(orgModule), eventpg.InboxWorkerConfig{
		Routines:       cfg.Kafka.Inbox.Routines,
		Slots:          cfg.Kafka.Inbox.Slots,
		Sleep:          cfg.Kafka.Inbox.Sleep,
		BatchSize:      cfg.Kafka.Inbox.BatchSize,
		MinNextAttempt: cfg.Kafka.Inbox.MinNextAttempt,
		MaxNextAttempt: cfg.Kafka.Inbox.MaxNextAttempt,
		MaxAttempts:    cfg.Kafka.Inbox.MaxAttempts,
	})

	run(func() { kafkaInboxArh.Start(ctx) })

	kafkaOutboxArch := messenger.NewOutbox(log, db, cfg.Kafka.Brokers, eventpg.OutboxWorkerConfig{
		Routines:       cfg.Kafka.Outbox.Routines,
		Slots:          cfg.Kafka.Inbox.Slots,
		Sleep:          cfg.Kafka.Inbox.Sleep,
		BatchSize:      cfg.Kafka.Inbox.BatchSize,
		MinNextAttempt: cfg.Kafka.Outbox.MinNextAttempt,
		MaxNextAttempt: cfg.Kafka.Outbox.MaxNextAttempt,
		MaxAttempts:    cfg.Kafka.Outbox.MaxAttempts,
	})

	run(func() { kafkaOutboxArch.Start(ctx) })
}
