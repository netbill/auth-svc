package responses

import (
	"github.com/netbill/auth-svc/internal/core/models"
	resources2 "github.com/netbill/auth-svc/pkg/resources"
)

func TokensPair(m models.TokensPair) resources2.TokensPair {
	resp := resources2.TokensPair{
		Data: resources2.TokensPairData{
			Id:   m.SessionID,
			Type: "tokens_pair",
			Attributes: resources2.TokensPairDataAttributes{
				AccessToken:  m.Access,
				RefreshToken: m.Refresh,
			},
		},
	}

	return resp
}
