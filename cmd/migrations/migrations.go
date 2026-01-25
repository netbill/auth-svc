package migrations

import (
	"context"
	"database/sql"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

//go:embed schema/*.sql
var Migrations embed.FS

var migrations = &migrate.EmbedFileSystemMigrationSource{
	FileSystem: Migrations,
	Root:       "schema",
}

func openDB(ctx context.Context, url string) (*pgxpool.Pool, *sql.DB, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create pgx pool")
	}
	db := stdlib.OpenDBFromPool(pool)
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		pool.Close()
		return nil, nil, errors.Wrap(err, "failed to ping database")
	}
	return pool, db, nil
}

func MigrateUp(ctx context.Context, url string) error {
	pool, db, err := openDB(ctx, url)
	if err != nil {
		return err
	}
	defer db.Close()
	defer pool.Close()

	applied, err := migrate.ExecContext(ctx, db, "postgres", migrations, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations (up)")
	}
	logrus.WithField("applied", applied).Info("migrations applied")

	return nil
}

func MigrateDown(ctx context.Context, url string) error {
	pool, db, err := openDB(ctx, url)
	if err != nil {
		return err
	}
	defer db.Close()
	defer pool.Close()

	applied, err := migrate.ExecContext(ctx, db, "postgres", migrations, migrate.Down)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations (down)")
	}
	logrus.WithField("applied", applied).Info("migrations applied")

	return nil
}
