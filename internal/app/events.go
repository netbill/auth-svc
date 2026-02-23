package app

import (
	"context"

	eventpg "github.com/netbill/eventbox/pg"
	"github.com/netbill/pgdbx"
)

func (a *App) CleanupInboxProcessingEvents(ctx context.Context, processIDs ...string) error {
	pool, err := a.config.PoolDB(ctx)
	if err != nil {
		a.log.WithError(err).Error("failed to connect to database")
		return err
	}
	defer pool.Close()

	db := pgdbx.NewDB(pool)
	err = eventpg.NewInbox(db).CleanProcessingInboxEvents(ctx, processIDs...)
	if err != nil {
		a.log.WithError(err).Error("failed to clean inbox processing")
		return err
	}

	a.log.Info("inbox processing cleaned successfully")
	return nil
}

func (a *App) CleanupOutboxProcessingEvents(ctx context.Context, processIDs ...string) error {
	pool, err := a.config.PoolDB(ctx)
	if err != nil {
		a.log.WithError(err).Error("failed to connect to database")
		return err
	}
	defer pool.Close()

	db := pgdbx.NewDB(pool)
	err = eventpg.NewOutbox(db).CleanProcessingOutboxEvents(ctx, processIDs...)
	if err != nil {
		a.log.WithError(err).Error("failed to clean outbox processing")
		return err
	}

	a.log.Info("outbox processing cleaned successfully")
	return nil
}

func (a *App) CleanupInboxFailedEvents(ctx context.Context) error {
	pool, err := a.config.PoolDB(ctx)
	if err != nil {
		a.log.WithError(err).Error("failed to connect to database")
		return err
	}
	defer pool.Close()

	db := pgdbx.NewDB(pool)
	err = eventpg.NewInbox(db).CleanFailedInboxEvents(ctx)
	if err != nil {
		a.log.WithError(err).Error("failed to clean inbox failed")
		return err
	}

	a.log.Info("inbox failed cleaned successfully")
	return nil
}

func (a *App) CleanupOutboxFailedEvents(ctx context.Context) error {
	pool, err := a.config.PoolDB(ctx)
	if err != nil {
		a.log.WithError(err).Error("failed to connect to database")
		return err
	}
	defer pool.Close()

	db := pgdbx.NewDB(pool)
	err = eventpg.NewOutbox(db).CleanFailedOutboxEvents(ctx)
	if err != nil {
		a.log.WithError(err).Error("failed to clean outbox failed")
		return err
	}

	a.log.Info("outbox failed cleaned successfully")
	return nil
}
