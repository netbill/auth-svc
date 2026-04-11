package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/netbill/auth-svc/internal/api/rest"
	"github.com/netbill/auth-svc/internal/api/rest/controller"
	"github.com/netbill/auth-svc/internal/api/rest/middlewares"
	"github.com/netbill/auth-svc/internal/modules/account"
	authmodule "github.com/netbill/auth-svc/internal/modules/auth"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/auth-svc/internal/passmanager"
	"github.com/netbill/auth-svc/internal/repo/chache"
	"github.com/netbill/auth-svc/internal/repo/pg"
	"github.com/netbill/auth-svc/internal/tokenmanager"
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

	// — database —

	pool, err := a.config.PoolDB(ctx)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	db := pgdbx.NewDB(pool)

	// — redis —

	redisClient := a.config.RedisClient()
	defer redisClient.Close()

	// — repos —

	accountRepo := pg.NewAccountRepo(db)
	emailRepo := pg.NewEmailRepo(db)
	passwordRepo := pg.NewPasswordRepo(db)
	sessionRepo := pg.NewSessionRepo(db)
	outboxRepo := pg.NewOutboxRepo(db, a.config.Kafka.Identity)

	// — caches —

	redisTTL := a.config.Redis.TTL

	accountCache := chache.NewAccountCache(redisClient, redisTTL.Account)
	emailCache := chache.NewEmailCache(redisClient, redisTTL.Email)
	passwordCache := chache.NewPasswordCache(redisClient, redisTTL.Password)
	sessionCache := chache.NewSessionCache(redisClient, redisTTL.Session)

	// — managers —

	passMgr := passmanager.New()

	tokenMgr := tokenmanager.New(tokenmanager.Config{
		Issuer:           a.config.Auth.Tokens.Issuer,
		AccessSecretKey:  a.config.Auth.Tokens.AccountAccess.SecretKey,
		AccessTTL:        a.config.Auth.Tokens.AccountAccess.TTL,
		RefreshSecretKey: a.config.Auth.Tokens.AccountRefresh.SecretKey,
		RefreshTTL:       a.config.Auth.Tokens.AccountRefresh.TTL,
		RefreshHashKey:   a.config.Auth.Tokens.AccountRefresh.HashKey,
	})

	// — module services —

	log := slog.Default()

	authSvc := authmodule.New(authmodule.ServiceDeps{
		AccountRepo:   accountRepo,
		SessionRepo:   sessionRepo,
		AccountCache:  accountCache,
		SessionsCache: sessionCache,
		Log:           log,
	})

	accountSvc := account.New(account.ServiceDeps{
		Auth:          authSvc,
		AccountRepo:   accountRepo,
		EmailRepo:     emailRepo,
		PasswordRepo:  passwordRepo,
		Tx:            db,
		AccountCache:  accountCache,
		EmailCache:    emailCache,
		PasswordCache: passwordCache,
		PassManager:   passMgr,
		Messenger:     outboxRepo,
		Log:           log,
	})

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
		Log:           log,
	})

	// — controllers —

	accountCtrl := controller.NewAccountController(accountSvc)
	sessionCtrl := controller.NewSessionController(sessionSvc, a.config.GoogleOAuth())

	// — rest server —

	mdll := middlewares.New(tokenMgr)
	router := rest.New(rest.ServerDeps{
		Accounts:    accountCtrl,
		Sessions:    sessionCtrl,
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

	// TODO: Kafka consumer (reads from topics, calls handlers directly — no InboxWorker)
	// Debezium reads WAL → outbox_events → Kafka (no OutboxWorker needed)

	a.log.Info("starting application")
	wg.Wait()
	return nil
}
