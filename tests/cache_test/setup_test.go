package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/internal/repo/chache"
	pkglog "github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

var (
	testCfg      *testutil.Config
	testPool     *pgxpool.Pool
	testRedis    *redis.Client
	testCacheTTL = 10 * time.Minute
	testLog      = pkglog.New("debug", "text", "test")
)

type noopMetrics struct{}

func (n *noopMetrics) UserCacheOp(_ context.Context, _ *error)     {}
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Database.Redis.Addr,
		Password: cfg.Database.Redis.Password,
		DB:       cfg.Database.Redis.DB,
	})
	testRedis = redisClient

	m.Run()

	_ = redisClient.Close()
	pool.Close()
}

func setupCacheTest(t *testing.T) *redis.Client {
	t.Helper()
	testutil.TruncateTables(t, testPool)
	testutil.FlushRedis(t, testRedis)
	return testRedis
}

func newUserCache(t *testing.T) *chache.UserCache {
	t.Helper()
	require.NotNil(t, testRedis)
	return chache.NewUserCache(testRedis, testCacheTTL, noop, testLog)
}

func newEmailCache(t *testing.T) *chache.EmailCache {
	t.Helper()
	require.NotNil(t, testRedis)
	return chache.NewEmailCache(testRedis, testCacheTTL, noop, testLog)
}

func newPasswordCache(t *testing.T) *chache.PasswordCache {
	t.Helper()
	require.NotNil(t, testRedis)
	return chache.NewPasswordCache(testRedis, testCacheTTL, noop, testLog)
}

func newSessionCache(t *testing.T) *chache.SessionCache {
	t.Helper()
	require.NotNil(t, testRedis)
	return chache.NewSessionCache(testRedis, testCacheTTL, noop, testLog)
}
