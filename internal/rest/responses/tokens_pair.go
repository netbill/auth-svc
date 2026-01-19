package responses

import (
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/resources"
)

func TokensPair(m models.TokensPair) resources.TokensPair {
	resp := resources.TokensPair{
		Data: resources.TokensPairData{
			Id:   m.SessionID,
			Type: "tokens_pair",
			Attributes: resources.TokensPairDataAttributes{
				AccessToken:  m.Access,
				RefreshToken: m.Refresh,
			},
		},
	}

	return resp
}
