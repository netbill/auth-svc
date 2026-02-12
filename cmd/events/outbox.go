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

func CleanupOutboxFailed(ctx context.Context, cfg config.Config, log *logium.Logger) error {
	pool, err := pgxpool.New(ctx, cfg.Database.SQL.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	db := pgdbx.NewDB(pool)

	OutboxCleaner := eventpg.NewOutboxCleaner(db)

	err = OutboxCleaner.CleanOutboxFailed(ctx)
	if err != nil {
		log.WithError(err).Error("failed to clean Outbox failed")
		return err
	}

	log.Info("Outbox failed cleaned successfully")
	return nil
}

func CleanupOutboxProcessing(ctx context.Context, cfg config.Config, log *logium.Logger, processIDs ...string) error {
	pool, err := pgxpool.New(ctx, cfg.Database.SQL.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	db := pgdbx.NewDB(pool)

	OutboxCleaner := eventpg.NewOutboxCleaner(db)

	err = OutboxCleaner.CleanOutboxProcessing(ctx, processIDs...)
	if err != nil {
		log.WithError(err).Error("failed to clean Outbox processing")
		return err
	}

	log.Info("Outbox processing cleaned successfully")
	return nil
}
