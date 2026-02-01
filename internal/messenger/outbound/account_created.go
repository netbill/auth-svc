package outbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/messenger/contracts"
	"github.com/netbill/evebox/header"
	"github.com/segmentio/kafka-go"
)

func (o *Outbound) WriteAccountCreated(
	ctx context.Context,
	account models.Account,
) error {
	payload, err := json.Marshal(contracts.AccountCreatedPayload{
		AccountID: account.ID,
		Username:  account.Username,
		Role:      account.Role,
		CreatedAt: account.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal account created payload, cause: %w", err)
	}

	event, err := o.outbox.CreateOutboxEvent(
		ctx,
		kafka.Message{
			Topic: contracts.AccountsTopicV1,
			Key:   []byte(account.ID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: header.EventID, Value: []byte(uuid.New().String())},
				{Key: header.EventType, Value: []byte(contracts.AccountCreatedEvent)},
				{Key: header.EventVersion, Value: []byte("1")},
				{Key: header.Producer, Value: []byte(contracts.AuthSvcGroup)},
				{Key: header.ContentType, Value: []byte("application/json")},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account created event, cause: %w", err)
	}

	o.log.Debugf("created outbox event %s for account %s, id %s", contracts.AccountCreatedEvent, event.ID.String(), account.ID.String())

	return nil
}
