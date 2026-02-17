package responses

import (
	"github.com/netbill/auth-svc/internal/core/models"
	resources2 "github.com/netbill/auth-svc/pkg/resources"
)

func Account(m models.Account) resources2.Account {
	resp := resources2.Account{
		Data: resources2.AccountData{
			Id:   m.ID,
			Type: "account",
			Attributes: resources2.AccountDataAttributes{
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
