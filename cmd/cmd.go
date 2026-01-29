package cmd

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/core/modules/organization"
	"github.com/netbill/auth-svc/internal/messenger"
	"github.com/netbill/auth-svc/internal/messenger/inbound"
	"github.com/netbill/auth-svc/internal/messenger/outbound"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/auth-svc/internal/repository/pg"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/rest/controller"
	"github.com/netbill/auth-svc/internal/rest/middlewares"
	"github.com/netbill/auth-svc/internal/tokenmanger"
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
)

func StartServices(ctx context.Context, cfg Config, log *logium.Logger, wg *sync.WaitGroup) {
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

	jwtTokenManager := tokenmanger.NewManager(tokenmanger.Config{
		AccessSK:   cfg.JWT.User.AccessToken.SecretKey,
		RefreshSK:  cfg.JWT.User.RefreshToken.SecretKey,
		RefreshHK:  cfg.JWT.User.RefreshToken.HashKey,
		AccessTTL:  cfg.JWT.User.AccessToken.TokenLifetime,
		RefreshTTL: cfg.JWT.User.RefreshToken.TokenLifetime,
		Iss:        cfg.Service.Name,
	})

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

	kafkaOutbound := outbound.New(log, db)

	accountCore := account.NewService(repo, jwtTokenManager, kafkaOutbound)
	orgCore := organization.New(repo)

	ctrl := controller.New(log, cfg.GoogleOAuth(), accountCore)
	mdll := middlewares.New(log, cfg.JWT.User.AccessToken.SecretKey)
	router := rest.New(log, mdll, ctrl)

	msgx := messenger.New(log, db, cfg.Kafka.Brokers...)

	run(func() {
		router.Run(ctx, rest.Config{
			Port:              cfg.Rest.Port,
			TimeoutRead:       cfg.Rest.Timeouts.Read,
			TimeoutReadHeader: cfg.Rest.Timeouts.ReadHeader,
			TimeoutWrite:      cfg.Rest.Timeouts.Write,
			TimeoutIdle:       cfg.Rest.Timeouts.Idle,
		})
	})

	log.Infof("starting kafka brokers %s", cfg.Kafka.Brokers)

	run(func() { msgx.RunProducer(ctx) })

	run(func() { msgx.RunConsumer(ctx, inbound.New(log, orgCore)) })

}
