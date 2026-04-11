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
	email models.AccountEmail,
) error {
	return r.write(
		ctx,
		evtypes.AccountsTopicV1,
		account.ID.String(),
		evtypes.AccountCreatedEvent,
		evtypes.AccountCreatedPayload{
			Account:      toEvAccount(account),
			AccountEmail: toEvAccountEmail(email),
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
			Account: toEvAccount(account),
		},
	)
}

func (r *OutboxRepo) WriteAccountDeleted(
	ctx context.Context,
	account models.Account,
	email models.AccountEmail,
) error {
	return r.write(
		ctx,
		evtypes.AccountsTopicV1,
		account.ID.String(),
		evtypes.AccountDeletedEvent,
		evtypes.AccountDeletedPayload{
			Account:      toEvAccount(account),
			AccountEmail: toEvAccountEmail(email),
		},
	)
}

func toEvAccount(a models.Account) evtypes.Account {
	return evtypes.Account{
		ID:        a.ID,
		Username:  a.Username,
		Role:      a.Role,
		Version:   a.Version,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		DeletedAt: a.DeletedAt,
	}
}

func toEvAccountEmail(e models.AccountEmail) evtypes.AccountEmail {
	return evtypes.AccountEmail{
		AccountID: e.AccountID,
		Email:     e.Email,
		Verified:  e.Verified,
		Version:   e.Version,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		DeletedAt: e.DeletedAt,
	}
}
