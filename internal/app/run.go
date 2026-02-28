package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/netbill/auth-svc/internal/core/modules/auth"
	"github.com/netbill/auth-svc/internal/core/modules/organization"
	"github.com/netbill/auth-svc/internal/messenger"
	"github.com/netbill/auth-svc/internal/messenger/handler"
	"github.com/netbill/auth-svc/internal/messenger/publisher"
	"github.com/netbill/auth-svc/internal/passmanager"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/auth-svc/internal/repository/pg"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/rest/controller"
	"github.com/netbill/auth-svc/internal/rest/middlewares"
	"github.com/netbill/auth-svc/internal/tokenmanager"
	"github.com/netbill/eventbox"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/netbill/pgdbx"
)

func (a *App) Run(ctx context.Context) error {
	var wg = &sync.WaitGroup{}

	run := func(f func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f()
		}()
	}

	pool, err := a.config.PoolDB(ctx)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	a.log.Info("starting application")

	db := pgdbx.NewDB(pool)

	repo := &repository.Repository{
		AccountsSql:      pg.NewAccountsQ(db),
		AccountEmailsSql: pg.NewAccountEmailsQ(db),
		AccountPassSql:   pg.NewAccountPasswordsQ(db),
		SessionsSql:      pg.NewSessionsQ(db),
		OrganizationsSql: pg.NewOrganizationsQ(db),
		OrgMembersSql:    pg.NewOrganizationMembersQ(db),
		TombstonesSql:    pg.NewTombstonesQ(db),
		TransactionSql:   db,
	}

	outbox := eventpg.NewOutbox(db)
	inbox := eventpg.NewInbox(db)

	producer := messenger.NewProducer(a.log, messenger.ProducerConfig{
		Producer: a.config.Kafka.Identity,
		Brokers:  a.config.Kafka.Brokers,
		AccountV1: messenger.ProduceKafkaConfig{
			RequiredAcks: a.config.Kafka.Produce.Topics.AccountsV1.RequiredAcks,
			Compression:  a.config.Kafka.Produce.Topics.AccountsV1.Compression,
			Balancer:     a.config.Kafka.Produce.Topics.AccountsV1.Balancer,
			BatchSize:    a.config.Kafka.Produce.Topics.AccountsV1.BatchSize,
			BatchTimeout: a.config.Kafka.Produce.Topics.AccountsV1.BatchTimeout,
		},
	})
	defer producer.Close()

	outbound := publisher.New(a.config.Kafka.Identity, outbox, producer)

	tokenManager := tokenmanager.New(tokenmanager.Config{
		Issuer:                  a.config.Auth.Tokens.Issuer,
		AccountAccessTTL:        a.config.Auth.Tokens.AccountAccess.TTL,
		AccountAccessSecretKey:  a.config.Auth.Tokens.AccountAccess.SecretKey,
		AccountRefreshTTL:       a.config.Auth.Tokens.AccountRefresh.TTL,
		AccountRefreshSecretKey: a.config.Auth.Tokens.AccountRefresh.SecretKey,
		AccountRefreshHashKey:   a.config.Auth.Tokens.AccountRefresh.HashKey,
	})

	authModule := auth.New(repo, tokenManager, outbound, passmanager.New())
	orgModule := organization.New(repo)

	ctrl := controller.New(authModule, a.config.GoogleOAuth())
	mdll := middlewares.New(tokenManager)
	router := rest.New(mdll, ctrl)

	run(func() {
		router.Run(ctx, a.log, rest.Config{
			Port:              a.config.Rest.Port,
			ReadTimeout:       a.config.Rest.Timeouts.Read,
			ReadHeaderTimeout: a.config.Rest.Timeouts.ReadHeader,
			WriteTimeout:      a.config.Rest.Timeouts.Write,
			IdleTimeout:       a.config.Rest.Timeouts.Idle,
		})
	})

	outboxWorker := messenger.NewOutboxWorker(a.log, outbox, producer, eventbox.OutboxWorkerConfig{
		Routines:       a.config.Kafka.Outbox.Routines,
		Slots:          a.config.Kafka.Outbox.Slots,
		BatchSize:      a.config.Kafka.Outbox.BatchSize,
		Sleep:          a.config.Kafka.Outbox.Sleep,
		MinNextAttempt: a.config.Kafka.Outbox.MinNextAttempt,
		MaxNextAttempt: a.config.Kafka.Outbox.MaxNextAttempt,
		MaxAttempts:    a.config.Kafka.Outbox.MaxAttempts,
	})
	defer outboxWorker.Clean()

	run(func() {
		outboxWorker.Run(ctx)
	})

	inbound := handler.New(a.log, orgModule)

	inboxWorker := messenger.NewInboxWorker(a.log, inbox, eventbox.InboxWorkerConfig{
		Routines:       a.config.Kafka.Inbox.Routines,
		Slots:          a.config.Kafka.Inbox.Slots,
		BatchSize:      a.config.Kafka.Inbox.BatchSize,
		Sleep:          a.config.Kafka.Inbox.Sleep,
		MinNextAttempt: a.config.Kafka.Inbox.MinNextAttempt,
		MaxNextAttempt: a.config.Kafka.Inbox.MaxNextAttempt,
		MaxAttempts:    a.config.Kafka.Inbox.MaxAttempts,
	}, inbound)
	defer inboxWorker.Clean()

	run(func() {
		inboxWorker.Run(ctx)
	})

	consumer := messenger.NewConsumer(a.log, inbox, messenger.ConsumerConfig{
		GroupID:    a.config.Kafka.Identity,
		Brokers:    a.config.Kafka.Brokers,
		MinBackoff: a.config.Kafka.Consume.Backoff.Min,
		MaxBackoff: a.config.Kafka.Consume.Backoff.Max,
		OrganizationsV1: messenger.ConsumeKafkaConfig{
			Instances:     a.config.Kafka.Consume.Topics.OrganizationsV1.Instances,
			MinBytes:      a.config.Kafka.Consume.Topics.OrganizationsV1.MinBytes,
			MaxBytes:      a.config.Kafka.Consume.Topics.OrganizationsV1.MaxBytes,
			MaxWait:       a.config.Kafka.Consume.Topics.OrganizationsV1.MaxWait,
			QueueCapacity: a.config.Kafka.Consume.Topics.OrganizationsV1.QueueCapacity,
		},
		OrgMembersV1: messenger.ConsumeKafkaConfig{
			Instances:     a.config.Kafka.Consume.Topics.OrgMembersV1.Instances,
			MinBytes:      a.config.Kafka.Consume.Topics.OrgMembersV1.MinBytes,
			MaxBytes:      a.config.Kafka.Consume.Topics.OrgMembersV1.MaxBytes,
			MaxWait:       a.config.Kafka.Consume.Topics.OrgMembersV1.MaxWait,
			QueueCapacity: a.config.Kafka.Consume.Topics.OrgMembersV1.QueueCapacity,
		},
	})
	defer consumer.Close()

	run(func() {
		consumer.Run(ctx)
	})

	wg.Wait()
	return nil
}
