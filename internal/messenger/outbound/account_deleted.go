package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/messenger/contracts"
	"github.com/netbill/evebox/header"
	"github.com/segmentio/kafka-go"
)

func (o *Outbound) WriteAccountDeleted(
	ctx context.Context,
	accountID uuid.UUID,
) error {
	payload, err := json.Marshal(contracts.AccountDeletedPayload{
		AccountID: accountID,
		DeletedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal account deleted payload, cause: %w", err)
	}

	event, err := o.outbox.CreateOutboxEvent(
		ctx,
		kafka.Message{
			Topic: contracts.AccountsTopicV1,
			Key:   []byte(accountID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: header.EventID, Value: []byte(uuid.New().String())}, // Outbox will fill this
				{Key: header.EventType, Value: []byte(contracts.AccountDeletedEvent)},
				{Key: header.EventVersion, Value: []byte("1")},
				{Key: header.Producer, Value: []byte(contracts.AuthSvcGroup)},
				{Key: header.ContentType, Value: []byte("application/json")},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account deleted event, cause: %w", err)
	}

	o.log.Debugf("created outbox event %s for account %s, id %s", contracts.AccountDeletedEvent, event.ID.String(), accountID.String())

	return err
}
