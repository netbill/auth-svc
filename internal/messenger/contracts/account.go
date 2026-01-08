package contracts

import "github.com/netbill/auth-svc/internal/core/models"

const AccountsTopicV1 = "accounts.v1"

const AccountCreatedEvent = "account.created"

type AccountCreatedPayload struct {
	Account models.Account `json:"account"`
	Email   string         `json:"email,omitempty"`
}

const AccountLoginEvent = "account.login"

type AccountLoginPayload struct {
	Account models.Account `json:"account"`
}

const AccountPasswordChangedEvent = "account.password.changed"

type AccountPasswordChangePayload struct {
	Account models.Account `json:"account"`
}

const AccountUsernameChangedEvent = "account.username.changed"

type AccountUsernameChangePayload struct {
	Account models.Account `json:"account"`
}

const AccountDeletedEvent = "account.deleted"

type AccountDeletedPayload struct {
	Account models.Account `json:"account"`
}
