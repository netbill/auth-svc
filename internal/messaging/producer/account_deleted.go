package producer

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/umisto/kafkakit/box"
	"github.com/umisto/kafkakit/header"
	"github.com/umisto/sso-svc/internal/domain/models"
	"github.com/umisto/sso-svc/internal/messaging/contracts"
)

func (s Service) AccountDeleted(
	ctx context.Context,
	account models.Account,
	email string,
) error {
	payload, err := json.Marshal(contracts.AccountDeletedPayload{
		Account: account,
		Email:   email,
	})
	if err != nil {
		return err
	}

	_, err = s.outbox.CreateOutboxEvent(
		ctx,
		box.OutboxStatusPending,
		kafka.Message{
			Topic: contracts.AccountsTopicV1,
			Key:   []byte(account.ID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: header.EventID, Value: []byte(uuid.New().String())}, // Outbox will fill this
				{Key: header.EventType, Value: []byte(contracts.AccountDeletedEvent)},
				{Key: header.EventVersion, Value: []byte("1")},
				{Key: header.Producer, Value: []byte(contracts.SsoSvcGroup)},
				{Key: header.ContentType, Value: []byte("application/json")},
			},
		},
	)

	return err
}
