package responses

import (
	"github.com/netbill/auth-svc/internal/core/models"
	resources2 "github.com/netbill/auth-svc/pkg/resources"
)

func AccountEmailData(ae models.AccountEmail) resources2.AccountEmail {
	return resources2.AccountEmail{
		Data: resources2.AccountEmailData{
			Id:   ae.AccountID,
			Type: "account_email",
			Attributes: resources2.AccountEmailDataAttributes{
				Email:     ae.Email,
				Version:   ae.Version,
				Verified:  ae.Verified,
				UpdatedAt: ae.UpdatedAt,
			},
		},
	}
}
