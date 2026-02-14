package evtypes

import (
	"time"

	"github.com/google/uuid"
)

const AccountsTopicV1 = "accounts.v1"

const AccountCreatedEvent = "account.created"

type AccountCreatedPayload struct {
	AccountID uuid.UUID `json:"account_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Version   int32     `json:"version"`

	CreatedAt time.Time `json:"created_at"`
}

const AccountUsernameUpdatedEvent = "account.username.updated"

type AccountUsernameUpdatedPayload struct {
	AccountID   uuid.UUID `json:"account_id"`
	NewUsername string    `json:"new_username"`
	Version     int32     `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const AccountDeletedEvent = "account.deleted"

type AccountDeletedPayload struct {
	AccountID uuid.UUID `json:"account_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
