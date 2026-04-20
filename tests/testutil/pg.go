package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/netbill/auth-svc/migrations"
	"github.com/netbill/pgdbx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/require"
)

// NewPoolGlobal connects to PG without *testing.T — for use in TestMain.
func NewPoolGlobal(url string) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), url)
}

// RunMigrationsGlobal applies migrations without *testing.T — for use in TestMain.
func RunMigrationsGlobal(pool *pgxpool.Pool) {
	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	applied, err := migrate.Exec(db, "postgres", migrations.Migrations, migrate.Up)
	if err != nil {
		panic(fmt.Sprintf("run migrations: %v", err))
	}
	_ = applied
}

func NewPool(t *testing.T, url string) *pgxpool.Pool {
	t.Helper()

	pool, err := pgxpool.New(context.Background(), url)
	require.NoError(t, err, "connect to test postgres")

	t.Cleanup(func() { pool.Close() })

	return pool
}

func NewDB(t *testing.T, url string) *pgdbx.DB {
	t.Helper()
	return pgdbx.NewDB(NewPool(t, url))
}

// RunMigrations applies all migrations to the test database.
func RunMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	applied, err := migrate.Exec(db, "postgres", migrations.Migrations, migrate.Up)
	require.NoError(t, err, "run migrations")
	t.Logf("migrations applied: %d", applied)
}

// TruncateTables clears all domain tables between tests.
func TruncateTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	conn, err := pool.Acquire(context.Background())
	require.NoError(t, err)
	defer conn.Release()

	_, err = conn.Exec(context.Background(), `
		TRUNCATE TABLE
			account_passwords,
			account_emails,
			sessions,
			outbox_events,
			accounts
		RESTART IDENTITY CASCADE
	`)
	require.NoError(t, err, "truncate tables")
}

// MustOpenStdDB returns a *sql.DB from pool (used for migrate).
func MustOpenStdDB(t *testing.T, pool *pgxpool.Pool) *sql.DB {
	t.Helper()
	db := stdlib.OpenDBFromPool(pool)
	t.Cleanup(func() { _ = db.Close() })
	return db
}
