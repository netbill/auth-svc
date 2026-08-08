package responses

import (
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/oapi"
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
) oapi.Account {
	res := &accountResponse{
		account: m,
	}
	for _, opt := range opts {
		opt(res)
	}

	data := oapi.AccountData{
		Id:   m.ID,
		Type: "account",
		Attributes: oapi.AccountDataAttributes{
			Role:      m.Role,
			Version:   m.Version,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
	}

	included := make([]oapi.AccountEmail, 0)
	if res.email != nil {
		included = append(included, AccountEmailData(*res.email))
	}

	return oapi.Account{
		Data:     data,
		Included: included,
	}
}

func AccountEmailData(ae models.AccountEmail) oapi.AccountEmail {
	return oapi.AccountEmail{
		Id:   ae.AccountID,
		Type: "account_email",
		Attributes: oapi.AccountEmailAttributes{
			Email:     ae.Email,
			Version:   ae.Version,
			Verified:  ae.Verified,
			UpdatedAt: ae.UpdatedAt,
		},
	}
}
