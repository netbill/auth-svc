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

func (r *OutboxRepo) WriteUserCreated(
	ctx context.Context,
	user models.User,
	email models.UserEmail,
) error {
	return r.write(
		ctx,
		evtypes.UsersTopicV1,
		user.ID.String(),
		evtypes.UserCreatedEvent,
		evtypes.UserCreatedPayload{
			User:      toEvUser(user),
			UserEmail: toEvUserEmail(email),
		},
	)
}

func (r *OutboxRepo) WriteUserUpdated(
	ctx context.Context,
	user models.User,
) error {
	return r.write(
		ctx,
		evtypes.UsersTopicV1,
		user.ID.String(),
		evtypes.UserUpdatedEvent,
		evtypes.UserUpdatedPayload{
			User: toEvUser(user),
		},
	)
}

func (r *OutboxRepo) WriteUserDeleted(
	ctx context.Context,
	user models.User,
	email models.UserEmail,
) error {
	return r.write(
		ctx,
		evtypes.UsersTopicV1,
		user.ID.String(),
		evtypes.UserDeletedEvent,
		evtypes.UserDeletedPayload{
			User:      toEvUser(user),
			UserEmail: toEvUserEmail(email),
		},
	)
}

func toEvUser(u models.User) evtypes.User {
	return evtypes.User{
		ID:        u.ID,
		Username:  u.Username,
		Pseudonym: u.Pseudonym,
		Bio:       u.Description,
		AvatarKey: u.AvatarKey,
		Version:   u.Version,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeletedAt: u.DeletedAt,
	}
}

func toEvUserEmail(e models.UserEmail) evtypes.UserEmail {
	return evtypes.UserEmail{
		UserID:    e.UserID,
		Email:     e.Email,
		Verified:  e.Verified,
		Version:   e.Version,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		DeletedAt: e.DeletedAt,
	}
}
