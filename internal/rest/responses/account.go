package responses

import (
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/resources"
)

func Account(m models.Account) resources.Account {
	resp := resources.Account{
		Data: resources.AccountData{
			Id:   m.ID,
			Type: "account",
			Attributes: resources.AccountDataAttributes{
				Role:      m.Role,
				Username:  m.Username,
				Version:   m.Version,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			},
		},
	}

	return resp
}
