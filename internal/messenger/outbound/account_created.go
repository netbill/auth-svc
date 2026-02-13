package outbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/messenger/evtypes"
	"github.com/netbill/eventbox/headers"
	"github.com/segmentio/kafka-go"
)

func (o *Outbound) WriteAccountCreated(
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

	_, err = o.outbox.WriteToOutbox(
		ctx,
		kafka.Message{
			Topic: evtypes.AccountsTopicV1,
			Key:   []byte(account.ID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: headers.EventID, Value: []byte(uuid.New().String())},
				{Key: headers.EventType, Value: []byte(evtypes.AccountCreatedEvent)},
				{Key: headers.EventVersion, Value: []byte("1")},
				{Key: headers.Producer, Value: []byte(evtypes.AuthSvcGroup)},
				{Key: headers.ContentType, Value: []byte("application/json")},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account created, cause: %w", err)
	}

	return nil
}
