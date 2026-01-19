package contracts

import (
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

const AccountsTopicV1 = "accounts.v1"

const AccountCreatedEvent = "account.created"

type AccountCreatedPayload struct {
	ID     uuid.UUID `json:"id"`
	Email  string    `json:"email,omitempty"`
	Role   string    `json:"role"`
	Status string    `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const AccountLoginEvent = "account.login"

type AccountLoginPayload struct {
	Account models.Account `json:"account"`
}

const AccountPasswordChangedEvent = "account.password.changed"

type AccountPasswordChangePayload struct {
	Account models.Account `json:"account"`
}

const AccountDeletedEvent = "account.deleted"

type AccountDeletedPayload struct {
	Account models.Account `json:"account"`
}
