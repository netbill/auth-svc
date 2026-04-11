package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/evtypes"
	"github.com/netbill/pgdbx"
)

const (
	outboxTable    = "outbox_events"
	outboxCols     = "event_id, topic, key, type, version, producer, payload"
	payloadVersion = 1
)

type OutboxRepo struct {
	db       *pgdbx.DB
	producer string
}

func NewOutboxRepo(db *pgdbx.DB, producer string) *OutboxRepo {
	return &OutboxRepo{db: db, producer: producer}
}

func (r *OutboxRepo) write(ctx context.Context, topic, key, eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	const query = `
		INSERT INTO ` + outboxTable + ` (` + outboxCols + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if _, err = r.db.Exec(ctx, query, uuid.New(), topic, key, eventType, payloadVersion, r.producer, data); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	return nil
}

func (r *OutboxRepo) WriteAccountCreated(
	ctx context.Context,
	account models.Account,
	_ models.AccountEmail,
) error {
	return r.write(
		ctx,
		evtypes.AccountsTopicV1,
		account.ID.String(),
		evtypes.AccountCreatedEvent,
		evtypes.AccountCreatedPayload{
			AccountID: account.ID,
			Username:  account.Username,
			Role:      account.Role,
			CreatedAt: account.CreatedAt,
		},
	)
}

func (r *OutboxRepo) WriteAccountUsernameUpdated(
	ctx context.Context,
	account models.Account,
) error {
	return r.write(
		ctx,
		evtypes.AccountsTopicV1,
		account.ID.String(),
		evtypes.AccountUsernameUpdatedEvent,
		evtypes.AccountUsernameUpdatedPayload{
			AccountID: account.ID,
			Username:  account.Username,
			Version:   account.Version,
			UpdatedAt: account.UpdatedAt,
		},
	)
}

func (r *OutboxRepo) WriteAccountDeleted(
	ctx context.Context,
	account models.Account,
	_ models.AccountEmail,
) error {
	var deletedAt = account.UpdatedAt
	if account.DeletedAt != nil {
		deletedAt = *account.DeletedAt
	}

	return r.write(
		ctx,
		evtypes.AccountsTopicV1,
		account.ID.String(),
		evtypes.AccountDeletedEvent,
		evtypes.AccountDeletedPayload{
			AccountID: account.ID,
			DeletedAt: deletedAt,
		},
	)
}
