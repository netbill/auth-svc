package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/eventbox"
	"github.com/netbill/evtypes"
)

func (p *Publisher) WriteAccountCreated(
	ctx context.Context,
	account models.Account,
) error {
	payload, err := json.Marshal(evtypes.AccountCreatedPayload{
		AccountID: account.ID,
		Username:  account.Username,
		Role:      account.Role,
		CreatedAt: account.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal account created payload, cause: %w", err)
	}

	_, err = p.outbox.WriteOutboxEvent(ctx, eventbox.Message{
		ID:       uuid.New(),
		Type:     evtypes.AccountCreatedEvent,
		Version:  1,
		Topic:    evtypes.AccountsTopicV1,
		Key:      account.ID.String(),
		Payload:  payload,
		Producer: p.identity,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account created, cause: %w", err)
	}

	return nil
}

func (p *Publisher) WriteAccountDeleted(
	ctx context.Context,
	accountID uuid.UUID,
) error {
	payload, err := json.Marshal(evtypes.AccountDeletedPayload{
		AccountID: accountID,
		DeletedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal account deleted payload, cause: %w", err)
	}

	_, err = p.outbox.WriteOutboxEvent(ctx, eventbox.Message{
		ID:       uuid.New(),
		Type:     evtypes.AccountDeletedEvent,
		Version:  1,
		Topic:    evtypes.AccountsTopicV1,
		Key:      accountID.String(),
		Payload:  payload,
		Producer: p.identity,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account deleted event, cause: %w", err)
	}

	return err
}

func (p *Publisher) WriteAccountUsernameUpdated(
	ctx context.Context,
	account models.Account,
) error {
	payload, err := json.Marshal(evtypes.AccountUsernameUpdatedPayload{
		AccountID:   account.ID,
		NewUsername: account.Username,
		Version:     account.Version,
		UpdatedAt:   account.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal account username updated payload, cause: %w", err)
	}

	_, err = p.outbox.WriteOutboxEvent(ctx, eventbox.Message{
		ID:       uuid.New(),
		Type:     evtypes.AccountUsernameUpdatedEvent,
		Version:  1,
		Topic:    evtypes.AccountsTopicV1,
		Key:      account.ID.String(),
		Payload:  payload,
		Producer: p.identity,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account username updated event, cause: %w", err)
	}

	return err
}
