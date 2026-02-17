package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	evtypes2 "github.com/netbill/auth-svc/pkg/evtypes"
	"github.com/netbill/eventbox/headers"
	"github.com/segmentio/kafka-go"
)

func (o *Outbound) WriteAccountCreated(
	ctx context.Context,
	account models.Account,
) error {
	payload, err := json.Marshal(evtypes2.AccountCreatedPayload{
		AccountID: account.ID,
		Username:  account.Username,
		Role:      account.Role,
		Version:   account.Version,
		CreatedAt: account.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal account created payload, cause: %w", err)
	}

	_, err = o.outbox.WriteToOutbox(
		ctx,
		kafka.Message{
			Topic: evtypes2.AccountsTopicV1,
			Key:   []byte(account.ID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: headers.EventID, Value: []byte(uuid.New().String())},
				{Key: headers.EventType, Value: []byte(evtypes2.AccountCreatedEvent)},
				{Key: headers.EventVersion, Value: []byte("1")},
				{Key: headers.Producer, Value: []byte(o.groupID)},
				{Key: headers.ContentType, Value: []byte("application/json")},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account created, cause: %w", err)
	}

	return nil
}

func (o *Outbound) WriteAccountDeleted(
	ctx context.Context,
	accountID uuid.UUID,
) error {
	payload, err := json.Marshal(evtypes2.AccountDeletedPayload{
		AccountID: accountID,
		DeletedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal account deleted payload, cause: %w", err)
	}

	_, err = o.outbox.WriteToOutbox(
		ctx,
		kafka.Message{
			Topic: evtypes2.AccountsTopicV1,
			Key:   []byte(accountID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: headers.EventID, Value: []byte(uuid.New().String())}, // Outbox will fill this
				{Key: headers.EventType, Value: []byte(evtypes2.AccountDeletedEvent)},
				{Key: headers.EventVersion, Value: []byte("1")},
				{Key: headers.Producer, Value: []byte(o.groupID)},
				{Key: headers.ContentType, Value: []byte("application/json")},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account deleted event, cause: %w", err)
	}

	return err
}

func (o *Outbound) WriteAccountUsernameUpdated(
	ctx context.Context,
	account models.Account,
) error {
	payload, err := json.Marshal(evtypes2.AccountUsernameUpdatedPayload{
		AccountID:   account.ID,
		NewUsername: account.Username,
		Version:     account.Version,
		UpdatedAt:   account.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal account username updated payload, cause: %w", err)
	}

	_, err = o.outbox.WriteToOutbox(
		ctx,
		kafka.Message{
			Topic: evtypes2.AccountsTopicV1,
			Key:   []byte(account.ID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: headers.EventID, Value: []byte(uuid.New().String())}, // Outbox will fill this
				{Key: headers.EventType, Value: []byte(evtypes2.AccountUsernameUpdatedEvent)},
				{Key: headers.EventVersion, Value: []byte("1")},
				{Key: headers.Producer, Value: []byte(o.groupID)},
				{Key: headers.ContentType, Value: []byte("application/json")},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account username updated event, cause: %w", err)
	}

	return err
}
