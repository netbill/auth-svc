package responses

import (
	"github.com/netbill/auth-svc/internal/domain/models"
	"github.com/netbill/auth-svc/resources"
)

func AccountEmailData(ae models.AccountEmail) resources.AccountEmail {
	return resources.AccountEmail{
		Data: resources.AccountEmailData{
			Id:   ae.AccountID,
			Type: resources.AccountEmailType,
			Attributes: resources.AccountEmailDataAttributes{
				Email:     ae.Email,
				Verified:  ae.Verified,
				UpdatedAt: ae.UpdatedAt,
			},
		},
	}
}
