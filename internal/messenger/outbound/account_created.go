package outbound

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/messenger/contracts"
	"github.com/netbill/evebox/header"
	"github.com/segmentio/kafka-go"
)

func (p Producer) WriteAccountCreated(
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
		return err
	}

	event, err := p.outbox.CreateOutboxEvent(
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
		return err
	}

	p.log.Debugf("created outbox event %s for account %s, id %s", contracts.AccountCreatedEvent, event.ID.String(), account.ID.String())

	return nil
}
