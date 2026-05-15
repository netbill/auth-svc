package repo_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/stretchr/testify/require"
)

var (
	testCfg  *testutil.Config
	testPool *pgxpool.Pool
)

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

	m.Run()

	pool.Close()
}

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	testutil.TruncateTables(t, testPool)
	return testPool
}

func requireConfig(t *testing.T) *testutil.Config {
	t.Helper()
	require.NotNil(t, testCfg)
	return testCfg
}
