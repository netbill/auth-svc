package contracts

import (
	"github.com/google/uuid"
)

const AccountsTopicV1 = "accounts.v1"

const AccountCreatedEvent = "account.created"

type AccountCreatedPayload struct {
	ID     uuid.UUID `json:"id"`
	Email  string    `json:"email,omitempty"`
	Role   string    `json:"role"`
	Status string    `json:"status"`
}

const AccountDeletedEvent = "account.deleted"

type AccountDeletedPayload struct {
	AccountID uuid.UUID `json:"account_id"`
}
