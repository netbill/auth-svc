package responses

import (
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/resources"
)

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
