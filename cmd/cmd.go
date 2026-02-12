package cmd

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/cmd/config"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/core/modules/organization"
	"github.com/netbill/auth-svc/internal/messenger"
	"github.com/netbill/auth-svc/internal/messenger/contracts"
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

func StartServices(ctx context.Context, cfg config.Config, log *logium.Logger, wg *sync.WaitGroup) {
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

	accountsSqlQ := pg.NewAccountsQ(db)
	accountEmailsSqlQ := pg.NewAccountEmailsQ(db)
	accountPasswordsSqlQ := pg.NewAccountPasswordsQ(db)
	sessionsSqlQ := pg.NewSessionsQ(db)
	orgMembersSqlQ := pg.NewOrganizationMembersQ(db)
	transactionerSql := pg.NewTransaction(db)

	repo := repository.NewRepository(
		transactionerSql,
		accountsSqlQ,
		accountEmailsSqlQ,
		accountPasswordsSqlQ,
		sessionsSqlQ,
		orgMembersSqlQ,
	)

	kafkaProducer := messenger.NewProducer(db, cfg.Kafka.Brokers...)
	kafkaOutbound := outbound.New(kafkaProducer)

	jwtTokenManager := tokenmanger.NewManager(tokenmanger.NewParams{
		AccessSK:   cfg.Auth.Account.Token.Access.SecretKey,
		RefreshSK:  cfg.Auth.Account.Token.Refresh.SecretKey,
		RefreshHK:  cfg.Auth.Account.Token.Refresh.HashKey,
		AccessTTL:  cfg.Auth.Account.Token.Access.Lifetime,
		RefreshTTL: cfg.Auth.Account.Token.Refresh.Lifetime,
	})

	authModule := account.NewService(repo, jwtTokenManager, kafkaOutbound)
	orgModule := organization.New(repo)

	responser := restkit.NewResponser()
	ctrl := controller.New(log, cfg.GoogleOAuth(), authModule, responser)
	mdll := middlewares.New(log, cfg.Auth.Account.Token.Access.SecretKey, responser)
	router := rest.New(log, mdll, ctrl)

	run(func() {
		router.Run(ctx, rest.Config{
			Port:              cfg.Rest.Port,
			TimeoutRead:       cfg.Rest.Timeouts.Read,
			TimeoutReadHeader: cfg.Rest.Timeouts.ReadHeader,
			TimeoutWrite:      cfg.Rest.Timeouts.Write,
			TimeoutIdle:       cfg.Rest.Timeouts.Idle,
		})
	})

	kafkaConsumer := messenger.NewConsumerArchitect(log, db, cfg.Kafka.Brokers, map[string]int{
		contracts.OrgMemberTopicV1: cfg.Kafka.Readers.OrgMemberV1,
	})

	run(func() { kafkaConsumer.Start(ctx) })

	kafkaInboxArh := messenger.NewInbox(log, db, inbound.New(orgModule), eventpg.InboxWorkerConfig{
		Routines:       cfg.Kafka.Inbox.Routines,
		MinSleep:       cfg.Kafka.Inbox.MinSleep,
		MaxSleep:       cfg.Kafka.Inbox.MaxSleep,
		MinBatch:       cfg.Kafka.Inbox.MinBatch,
		MaxBatch:       cfg.Kafka.Inbox.MaxBatch,
		MinNextAttempt: cfg.Kafka.Inbox.MinNextAttempt,
		MaxNextAttempt: cfg.Kafka.Inbox.MaxNextAttempt,
		MaxAttempts:    cfg.Kafka.Inbox.MaxAttempts,
	})

	run(func() { kafkaInboxArh.Start(ctx) })

	kafkaOutboxArch := messenger.NewOutbox(log, db, cfg.Kafka.Brokers, eventpg.OutboxWorkerConfig{
		Routines:       cfg.Kafka.Outbox.Routines,
		MinSleep:       cfg.Kafka.Outbox.MinSleep,
		MaxSleep:       cfg.Kafka.Outbox.MaxSleep,
		MinBatch:       cfg.Kafka.Outbox.MinBatch,
		MaxBatch:       cfg.Kafka.Outbox.MaxBatch,
		MinNextAttempt: cfg.Kafka.Outbox.MinNextAttempt,
		MaxNextAttempt: cfg.Kafka.Outbox.MaxNextAttempt,
		MaxAttempts:    cfg.Kafka.Outbox.MaxAttempts,
	})

	run(func() { kafkaOutboxArch.Start(ctx) })
}
