package responses

import (
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/resources"
)

type accountResponse struct {
	account models.Account
	email   *models.AccountEmail
}

type AccountOption func(*accountResponse)

func WithAccountEmail(email models.AccountEmail) AccountOption {
	return func(res *accountResponse) {
		res.email = &email
	}
}

func Account(
	m models.Account,
	opts ...AccountOption,
) resources.Account {
	res := &accountResponse{
		account: m,
	}
	for _, opt := range opts {
		opt(res)
	}

	data := resources.AccountData{
		Id:   m.ID,
		Type: "account",
		Attributes: resources.AccountDataAttributes{
			Role:      m.Role,
			Username:  m.Username,
			Version:   m.Version,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
	}

	included := make([]resources.AccountEmail, 0)
	if res.email != nil {
		included = append(included, AccountEmailData(*res.email))
	}

	return resources.Account{
		Data:     data,
		Included: included,
	}
}

func AccountEmailData(ae models.AccountEmail) resources.AccountEmail {
	return resources.AccountEmail{
		Id:   ae.AccountID,
		Type: "account_email",
		Attributes: resources.AccountEmailAttributes{
			Email:     ae.Email,
			Version:   ae.Version,
			Verified:  ae.Verified,
			UpdatedAt: ae.UpdatedAt,
		},
	}
}
