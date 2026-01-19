package responses

import (
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/resources"
)

func AccountEmailData(ae models.AccountEmail) resources.AccountEmail {
	return resources.AccountEmail{
		Data: resources.AccountEmailData{
			Id:   ae.AccountID,
			Type: "account_email",
			Attributes: resources.AccountEmailDataAttributes{
				Email:     ae.Email,
				Verified:  ae.Verified,
				UpdatedAt: ae.UpdatedAt,
			},
		},
	}
}
