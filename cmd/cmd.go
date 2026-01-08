package cmd

import (
	"context"
	"database/sql"
	"sync"

	"github.com/netbill/auth-svc/internal"
	"github.com/netbill/auth-svc/internal/core/modules/auth"
	"github.com/netbill/auth-svc/internal/messenger"
	"github.com/netbill/auth-svc/internal/messenger/outbound"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/rest/controller"
	"github.com/netbill/auth-svc/internal/token"
	"github.com/netbill/evebox/box/outbox"
	"github.com/netbill/logium"
	"github.com/netbill/restkit/mdlv"
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

	outBox := outbox.New(pg)

	jwtTokenManager := token.NewManager(token.Config{
		AccessSK:   cfg.JWT.User.AccessToken.SecretKey,
		RefreshSK:  cfg.JWT.User.RefreshToken.SecretKey,
		AccessTTL:  cfg.JWT.User.AccessToken.TokenLifetime,
		RefreshTTL: cfg.JWT.User.RefreshToken.TokenLifetime,
		Iss:        cfg.Service.Name,
	})

	kafkaOutbound := outbound.New(log, outBox)

	core := auth.NewService(repo, jwtTokenManager, kafkaOutbound)

	ctrl := controller.New(log, cfg.GoogleOAuth(), core)

	mdll := mdlv.New(cfg.JWT.User.AccessToken.SecretKey, rest.AccountDataCtxKey)
	router := rest.New(log, mdll, ctrl)

	kafkaProducer := messenger.NewProducer(log, outBox, cfg.Kafka.Brokers...)

	log.Infof("starting kafka brokers %s", cfg.Kafka.Brokers)

	run(func() { router.Run(ctx, cfg) })

	run(func() { kafkaProducer.Run(ctx) })

}
