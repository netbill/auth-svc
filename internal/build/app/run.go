package app

import (
	"context"
	"fmt"
	"sync"

	grpcapi "github.com/netbill/auth-svc/internal/api/grpc"
	"github.com/netbill/auth-svc/internal/api/rest"
	"github.com/netbill/auth-svc/internal/api/rest/controller"
	"github.com/netbill/auth-svc/internal/api/rest/middlewares"
	"github.com/netbill/auth-svc/internal/bus"
	"github.com/netbill/auth-svc/internal/modules/account"
	authmodule "github.com/netbill/auth-svc/internal/modules/auth"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/auth-svc/internal/observability/metrics"
	"github.com/netbill/auth-svc/internal/observability/telemetry"
	"github.com/netbill/auth-svc/internal/repo/chache"
	"github.com/netbill/auth-svc/internal/repo/pg"
	"github.com/netbill/auth-svc/pkg/googleid"
	"github.com/netbill/auth-svc/pkg/passmanager"
	"github.com/netbill/auth-svc/pkg/tokenmanager"
	"github.com/netbill/pgdbx"
)

func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	run := func(f func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f()
		}()
	}

	shutdownTelemetry, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:       a.config.Service.Name,
		CollectorEndpoint: a.config.OTEL.CollectorEndpoint,
		SamplingRatio:     a.config.OTEL.SamplingRatio,
	})
	if err != nil {
		return fmt.Errorf("init telemetry: %w", err)
	}
	defer shutdownTelemetry(ctx)

	svcMetrics, err := metrics.New()
	if err != nil {
		return fmt.Errorf("init metrics: %w", err)
	}

	pool, err := a.config.PoolDB(ctx)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	db := pgdbx.NewDB(pool)
	redisClient := a.config.RedisClient()
	defer redisClient.Close()

	accountRepo := pg.NewAccountRepo(db)
	emailRepo := pg.NewEmailRepo(db)
	passwordRepo := pg.NewPasswordRepo(db)
	sessionRepo := pg.NewSessionRepo(db)
	outboxRepo := pg.NewOutboxRepo(db, a.config.Kafka.Identity)

	redisTTL := a.config.Database.Redis.TTL

	accountCache := chache.NewAccountCache(redisClient, redisTTL.Account, svcMetrics, a.log)
	emailCache := chache.NewEmailCache(redisClient, redisTTL.Email, svcMetrics, a.log)
	passwordCache := chache.NewPasswordCache(redisClient, redisTTL.Password, svcMetrics, a.log)
	sessionCache := chache.NewSessionCache(redisClient, redisTTL.Session, svcMetrics, a.log)
	qrCache := chache.NewQRCache(redisClient)

	qrPublisher := bus.NewPublisher(redisClient)
	qrSubscriber := bus.NewSubscriber(redisClient)

	passMgr := passmanager.New(a.config.Auth.PassBcryptCost)

	tokenMgr := tokenmanager.New(tokenmanager.Config{
		Issuer:           a.config.Auth.Tokens.Issuer,
		AccessSecretKey:  a.config.Auth.Tokens.AccountAccess.SecretKey,
		AccessTTL:        a.config.Auth.Tokens.AccountAccess.TTL,
		RefreshSecretKey: a.config.Auth.Tokens.AccountRefresh.SecretKey,
		RefreshTTL:       a.config.Auth.Tokens.AccountRefresh.TTL,
		RefreshHashKey:   a.config.Auth.Tokens.AccountRefresh.HashKey,
	})

	authSvc := authmodule.New(authmodule.ServiceDeps{
		AccountRepo: accountRepo,
		SessionRepo: sessionRepo,
	})

	accountSvc := account.New(account.ServiceDeps{
		Auth:          authSvc,
		AccountRepo:   accountRepo,
		EmailRepo:     emailRepo,
		PasswordRepo:  passwordRepo,
		SessionRepo:   sessionRepo,
		Tx:            db,
		AccountCache:  accountCache,
		EmailCache:    emailCache,
		PasswordCache: passwordCache,
		SessionsCache: sessionCache,
		PassManager:   passMgr,
		Messenger:     outboxRepo,
	})

	broker := bus.NewBroker(qrPublisher, qrSubscriber)

	sessionSvc := session.New(session.ServiceDeps{
		Auth:          authSvc,
		AccountRepo:   accountRepo,
		EmailRepo:     emailRepo,
		PasswordRepo:  passwordRepo,
		SessionRepo:   sessionRepo,
		Tx:            db,
		PasswordCache: passwordCache,
		AccountCache:  accountCache,
		SessionsCache: sessionCache,
		PassManager:   passMgr,
		TokenManager:  tokenMgr,
		QRStore:       qrCache,
		Bus:           broker,
	})

	accountCtrl := controller.NewAccountController(accountSvc, svcMetrics)
	sessionCtrl := controller.NewSessionController(sessionSvc, a.config.GoogleOAuth(), svcMetrics, broker)

	mdll := middlewares.New(tokenMgr)
	router := rest.New(rest.ServerDeps{
		Accounts:    accountCtrl,
		Sessions:    sessionCtrl,
		QR:          sessionCtrl,
		Middlewares: mdll,
		Log:         a.log,
	})

	run(func() {
		router.Run(ctx, rest.Config{
			Port:              a.config.Rest.Port,
			ReadTimeout:       a.config.Rest.Timeouts.Read,
			ReadHeaderTimeout: a.config.Rest.Timeouts.ReadHeader,
			WriteTimeout:      a.config.Rest.Timeouts.Write,
			IdleTimeout:       a.config.Rest.Timeouts.Idle,
		})
	})

	googleVerifier := googleid.New(a.config.Auth.OAuth.Google.ClientID)

	grpcServer := grpcapi.New(grpcapi.ServerDeps{
		Auth:     authSvc,
		Accounts: accountSvc,
		Sessions: sessionSvc,
		Google:   googleVerifier,
		Metrics:  svcMetrics,
		TokenMgr: tokenMgr,
		Log:      a.log,
	})

	run(func() {
		grpcServer.Run(ctx, grpcapi.Config{
			Port: a.config.GRPC.Port,
		})
	})

	a.log.Info("starting application")
	wg.Wait()
	return nil
}
