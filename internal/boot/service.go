package boot

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/internal/core/modules/auth"
	"github.com/netbill/auth-svc/internal/core/modules/organization"
	"github.com/netbill/auth-svc/internal/messenger"
	"github.com/netbill/auth-svc/internal/messenger/inbound"
	"github.com/netbill/auth-svc/internal/messenger/outbound"
	"github.com/netbill/auth-svc/internal/passmanager"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/auth-svc/internal/repository/pg"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/rest/controller"
	"github.com/netbill/auth-svc/internal/rest/middlewares"
	"github.com/netbill/auth-svc/internal/tokenmanger"
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
	"github.com/netbill/restkit"
)

func RunService(ctx context.Context, log *logium.Entry, wg *sync.WaitGroup, cfg *Config) {
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

	msg := messenger.NewManager(log, db, cfg.Kafka)

	kafkaProducer := msg.NewProducer()
	kafkaOutbound := outbound.New(kafkaProducer)

	tokenManager := tokenmanger.New(ServiceName, cfg.Auth.Tokens)

	authModule := auth.New(repo, tokenManager, kafkaOutbound, passmanager.New())
	orgModule := organization.New(repo)

	responser := restkit.NewResponser()
	ctrl := controller.New(authModule, cfg.GoogleOAuth(), responser)
	mdll := middlewares.New(tokenManager, responser)
	router := rest.New(mdll, ctrl)

	run(func() {
		router.Run(ctx, log, cfg.Rest)
	})

	run(func() { msg.RunInbox(ctx, inbound.New(orgModule)) })

	run(func() { msg.RunConsumer(ctx) })

	run(func() { msg.RunOutbox(ctx) })
}
