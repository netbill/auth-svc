package outbound

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/messenger/contracts"
	"github.com/netbill/evebox/header"
	"github.com/segmentio/kafka-go"
)

func (p Producer) WriteAccountDeleted(
	ctx context.Context,
	accountID uuid.UUID,
) error {
	payload, err := json.Marshal(contracts.AccountDeletedPayload{
		AccountID: accountID,
	})
	if err != nil {
		return err
	}

	eventID := uuid.New().String()
	_, err = p.outbox.CreateOutboxEvent(
		ctx,
		kafka.Message{
			Topic: contracts.AccountsTopicV1,
			Key:   []byte(accountID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: header.EventID, Value: []byte(eventID)}, // Outbox will fill this
				{Key: header.EventType, Value: []byte(contracts.AccountDeletedEvent)},
				{Key: header.EventVersion, Value: []byte("1")},
				{Key: header.Producer, Value: []byte(contracts.AuthSvcGroup)},
				{Key: header.ContentType, Value: []byte("application/json")},
			},
		},
	)

	p.log.Debugf("created outbox event %s for account %s, id %s", contracts.AccountDeletedEvent, eventID, accountID.String())

	return err
}
