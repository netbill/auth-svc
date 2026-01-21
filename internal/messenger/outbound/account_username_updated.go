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

func (p Producer) WriteAccountUsernameUpdated(
	ctx context.Context,
	account models.Account,
) error {
	payload, err := json.Marshal(contracts.AccountUsernameUpdatedPayload{
		AccountID:   account.ID,
		NewUsername: account.Username,
		UpdatedAt:   account.UpdatedAt,
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
				{Key: header.EventID, Value: []byte(uuid.New().String())}, // Outbox will fill this
				{Key: header.EventType, Value: []byte(contracts.AccountUsernameUpdatedEvent)},
				{Key: header.EventVersion, Value: []byte("1")},
				{Key: header.Producer, Value: []byte(contracts.AuthSvcGroup)},
				{Key: header.ContentType, Value: []byte("application/json")},
			},
		},
	)
	if err != nil {
		return err
	}

	p.log.Debugf("created outbox event %s for account %s, id %s", contracts.AccountUsernameUpdatedEvent, event.ID.String(), account.ID.String())

	return err
}
