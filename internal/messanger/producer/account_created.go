package producer

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/domain/models"
	"github.com/netbill/auth-svc/internal/messanger/contracts"
	"github.com/netbill/kafkakit/box"
	"github.com/netbill/kafkakit/header"
	"github.com/segmentio/kafka-go"
)

func (s Service) WriteAccountCreated(
	ctx context.Context,
	account models.Account,
	email string,
) error {
	payload, err := json.Marshal(contracts.AccountCreatedPayload{
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
				{Key: header.EventType, Value: []byte(contracts.AccountCreatedEvent)},
				{Key: header.EventVersion, Value: []byte("1")},
				{Key: header.Producer, Value: []byte(contracts.SsoSvcGroup)},
				{Key: header.ContentType, Value: []byte("application/json")},
			},
		},
	)

	return err
}
