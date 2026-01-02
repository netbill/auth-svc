package responses

import (
	"github.com/netbill/auth-svc/internal/domain/models"
	"github.com/netbill/auth-svc/resources"
)

func TokensPair(m models.TokensPair) resources.TokensPair {
	resp := resources.TokensPair{
		Data: resources.TokensPairData{
			Id:   m.SessionID,
			Type: resources.TokensPairType,
			Attributes: resources.TokensPairDataAttributes{
				AccessToken:  m.Access,
				RefreshToken: m.Refresh,
			},
		},
	}

	return resp
}
