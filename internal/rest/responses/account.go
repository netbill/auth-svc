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
				Status:    m.Status,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			},
		},
	}

	return resp
}
