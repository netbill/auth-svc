package events

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/cmd/config"
	eventpg "github.com/netbill/eventbox/pg"
	"github.com/netbill/logium"
	"github.com/netbill/pgdbx"
)

func CleanupInboxFailed(ctx context.Context, cfg config.Config, log *logium.Logger) error {
	pool, err := pgxpool.New(ctx, cfg.Database.SQL.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	db := pgdbx.NewDB(pool)

	inboxCleaner := eventpg.NewInboxCleaner(db)

	err = inboxCleaner.CleanInboxFailed(ctx)
	if err != nil {
		log.WithError(err).Error("failed to clean inbox failed")
		return err
	}

	log.Info("inbox failed cleaned successfully")
	return nil
}

func CleanupInboxProcessing(ctx context.Context, cfg config.Config, log *logium.Logger, processIDs ...string) error {
	pool, err := pgxpool.New(ctx, cfg.Database.SQL.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	db := pgdbx.NewDB(pool)

	inboxCleaner := eventpg.NewInboxCleaner(db)

	err = inboxCleaner.CleanInboxProcessing(ctx, processIDs...)
	if err != nil {
		log.WithError(err).Error("failed to clean inbox processing")
		return err
	}

	log.Info("inbox processing cleaned successfully")
	return nil
}
