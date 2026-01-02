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

func (s Service) WriteAccountLogin(
	ctx context.Context,
	account models.Account,
) error {
	payload, err := json.Marshal(contracts.AccountLoginPayload{
		Account: account,
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
				{Key: header.EventType, Value: []byte(contracts.AccountLoginEvent)},
				{Key: header.EventVersion, Value: []byte("1")},
				{Key: header.Producer, Value: []byte(contracts.SsoSvcGroup)},
				{Key: header.ContentType, Value: []byte("application/json")},
			},
		},
	)

	return err
}
