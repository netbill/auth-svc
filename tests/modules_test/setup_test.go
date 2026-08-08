package modules_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
	authmodule "github.com/netbill/auth-svc/internal/modules/auth"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/auth-svc/internal/repo/chache"
	"github.com/netbill/auth-svc/internal/repo/pg"
	pkglog "github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/auth-svc/pkg/passmanager"
	"github.com/netbill/auth-svc/pkg/tokenmanager"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/netbill/pgdbx"
	"github.com/redis/go-redis/v9"
)

var (
	testCfg   *testutil.Config
	testPool  *pgxpool.Pool
	testRedis *redis.Client
	testLog   = pkglog.New("debug", "text", "test")
)

type noopMetrics struct{}

func (n *noopMetrics) AccountCacheOp(_ context.Context, _ *error)  {}
func (n *noopMetrics) EmailCacheOp(_ context.Context, _ *error)    {}
func (n *noopMetrics) PasswordCacheOp(_ context.Context, _ *error) {}
func (n *noopMetrics) SessionCacheOp(_ context.Context, _ *error)  {}

var noop = &noopMetrics{}

func TestMain(m *testing.M) {
	cfg, err := testutil.LoadConfig("test_config.yaml")
	if err != nil {
		panic("load test config: " + err.Error())
	}
	testCfg = cfg

	pool, err := testutil.NewPoolGlobal(cfg.Database.SQL.URL)
	if err != nil {
		panic("connect to postgres: " + err.Error())
	}
	testPool = pool
	testutil.RunMigrationsGlobal(pool)

	testRedis = redis.NewClient(&redis.Options{
		Addr:     cfg.Database.Redis.Addr,
		Password: cfg.Database.Redis.Password,
		DB:       cfg.Database.Redis.DB,
	})

	m.Run()

	_ = testRedis.Close()
	pool.Close()
}

func setup(t *testing.T) (*pgdbx.DB, *redis.Client) {
	t.Helper()
	testutil.TruncateTables(t, testPool)
	testutil.FlushRedis(t, testRedis)
	return pgdbx.NewDB(testPool), testRedis
}

const cacheTTL = 10 * time.Minute

func newServices(t *testing.T, db *pgdbx.DB, rc *redis.Client) (*account.Service, *session.Service) {
	t.Helper()

	accountRepo := pg.NewAccountRepo(db)
	emailRepo := pg.NewEmailRepo(db)
	passwordRepo := pg.NewPasswordRepo(db)
	sessionRepo := pg.NewSessionRepo(db)

	accountCache := chache.NewAccountCache(rc, cacheTTL, noop, testLog)
	emailCache := chache.NewEmailCache(rc, cacheTTL, noop, testLog)
	passwordCache := chache.NewPasswordCache(rc, cacheTTL, noop, testLog)
	sessionCache := chache.NewSessionCache(rc, cacheTTL, noop, testLog)

	authSvc := authmodule.New(authmodule.ServiceDeps{
		AccountRepo: accountRepo,
		SessionRepo: sessionRepo,
	})

	passMgr := passmanager.New(testCfg.Auth.PassBcryptCost)

	tokenMgr := tokenmanager.New(tokenmanager.Config{
		Issuer:           testCfg.Auth.Tokens.Issuer,
		AccessSecretKey:  testCfg.Auth.Tokens.AccountAccess.SecretKey,
		AccessTTL:        testCfg.Auth.Tokens.AccountAccess.TTL,
		RefreshSecretKey: testCfg.Auth.Tokens.AccountRefresh.SecretKey,
		RefreshTTL:       testCfg.Auth.Tokens.AccountRefresh.TTL,
		RefreshHashKey:   testCfg.Auth.Tokens.AccountRefresh.HashKey,
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
		Messenger:     &noopMessenger{},
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
	})

	return accountSvc, sessionSvc
}

type noopMessenger struct{}

func (n *noopMessenger) WriteAccountCreated(_ context.Context, _ models.Account, _ models.AccountEmail) error {
	return nil
}

func (n *noopMessenger) WriteAccountDeleted(_ context.Context, _ models.Account, _ models.AccountEmail) error {
	return nil
}
