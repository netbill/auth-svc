package cmd

import (
	"context"
	"database/sql"
	"sync"

	"github.com/netbill/auth-svc/internal"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/core/modules/organization"
	"github.com/netbill/auth-svc/internal/messenger"
	"github.com/netbill/auth-svc/internal/messenger/inbound"
	"github.com/netbill/auth-svc/internal/messenger/outbound"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/rest/controller"
	"github.com/netbill/auth-svc/internal/rest/middlewares"
	"github.com/netbill/auth-svc/internal/tokenmanger"
	"github.com/netbill/logium"
)

func StartServices(ctx context.Context, cfg internal.Config, log logium.Logger, wg *sync.WaitGroup) {
	run := func(f func()) {
		wg.Add(1)
		go func() {
			f()
			wg.Done()
		}()
	}

	pg, err := sql.Open("postgres", cfg.Database.SQL.URL)
	if err != nil {
		log.Fatal("failed to connect to database", "error", err)
	}

	repo := repository.New(pg)

	jwtTokenManager := tokenmanger.NewManager(tokenmanger.Config{
		AccessSK:   cfg.JWT.User.AccessToken.SecretKey,
		RefreshSK:  cfg.JWT.User.RefreshToken.SecretKey,
		RefreshHK:  cfg.JWT.User.RefreshToken.HashKey,
		AccessTTL:  cfg.JWT.User.AccessToken.TokenLifetime,
		RefreshTTL: cfg.JWT.User.RefreshToken.TokenLifetime,
		Iss:        cfg.Service.Name,
	})

	kafkaOutbound := outbound.New(log, pg)

	accountCore := account.NewService(repo, jwtTokenManager, kafkaOutbound)
	orgCore := organization.New(repo)

	ctrl := controller.New(log, cfg.GoogleOAuth(), accountCore)
	mdll := middlewares.New(log, cfg.JWT.User.AccessToken.SecretKey)
	router := rest.New(log, mdll, ctrl)

	msgx := messenger.New(log, pg, cfg.Kafka.Brokers...)

	run(func() { router.Run(ctx, cfg) })

	log.Infof("starting kafka brokers %s", cfg.Kafka.Brokers)

	run(func() { msgx.RunProducer(ctx) })

	run(func() { msgx.RunConsumer(ctx, inbound.New(log, orgCore)) })

}
