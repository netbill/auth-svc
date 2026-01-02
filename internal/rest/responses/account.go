package responses

import (
	"github.com/netbill/auth-svc/internal/domain/models"
	"github.com/netbill/auth-svc/resources"
)

func Account(m models.Account) resources.Account {
	resp := resources.Account{
		Data: resources.AccountData{
			Id:   m.ID,
			Type: resources.AccountType,
			Attributes: resources.AccountDataAttributes{
				Username:  m.Username,
				Role:      m.Role,
				Status:    m.Status,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			},
		},
	}

	return resp
}
