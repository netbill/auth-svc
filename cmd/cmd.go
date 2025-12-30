package cmd

import (
	"context"
	"database/sql"
	"sync"

	"github.com/umisto/kafkakit/box"
	"github.com/umisto/logium"
	"github.com/umisto/sso-svc/internal"
	"github.com/umisto/sso-svc/internal/domain/modules/auth"
	"github.com/umisto/sso-svc/internal/messaging/producer"
	"github.com/umisto/sso-svc/internal/repository"
	"github.com/umisto/sso-svc/internal/rest"
	"github.com/umisto/sso-svc/internal/rest/controller"
	"github.com/umisto/sso-svc/internal/rest/middlewares"
	"github.com/umisto/sso-svc/internal/token"
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

	repository := repository.New(pg)

	kafkaBox := box.New(pg)

	jwtTokenManager := token.NewManager(token.Config{
		AccessSK:   cfg.JWT.User.AccessToken.SecretKey,
		RefreshSK:  cfg.JWT.User.RefreshToken.SecretKey,
		AccessTTL:  cfg.JWT.User.AccessToken.TokenLifetime,
		RefreshTTL: cfg.JWT.User.RefreshToken.TokenLifetime,
		Iss:        cfg.Service.Name,
	})

	kafkaProducer := producer.New(log, cfg.Kafka.Brokers, kafkaBox)

	core := auth.NewService(repository, jwtTokenManager, kafkaProducer)

	ctrl := controller.New(log, cfg.GoogleOAuth(), core)
	mdlv := middlewares.New(log)

	run(func() { rest.Run(ctx, cfg, log, mdlv, ctrl) })

	run(func() { kafkaProducer.Run(ctx) })
}
