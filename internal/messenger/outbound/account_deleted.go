package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/messenger/evtypes"
	"github.com/netbill/eventbox/headers"
	"github.com/segmentio/kafka-go"
)

func (o *Outbound) WriteAccountDeleted(
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

	_, err = o.outbox.WriteToOutbox(
		ctx,
		kafka.Message{
			Topic: evtypes.AccountsTopicV1,
			Key:   []byte(accountID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: headers.EventID, Value: []byte(uuid.New().String())}, // Outbox will fill this
				{Key: headers.EventType, Value: []byte(evtypes.AccountDeletedEvent)},
				{Key: headers.EventVersion, Value: []byte("1")},
				{Key: headers.Producer, Value: []byte(evtypes.AuthSvcGroup)},
				{Key: headers.ContentType, Value: []byte("application/json")},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox event for account deleted event, cause: %w", err)
	}

	return err
}
