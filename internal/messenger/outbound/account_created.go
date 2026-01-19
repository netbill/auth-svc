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
	email string,
) error {
	payload, err := json.Marshal(contracts.AccountCreatedPayload{
		ID:        account.ID,
		Email:     email,
		Role:      account.Role,
		Status:    account.Status,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
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

	p.log.Debugf("created outbox event %s for account %s, id %s", contracts.AccountCreatedEvent, account.ID.String(), event.ID.String())

	return err
}
