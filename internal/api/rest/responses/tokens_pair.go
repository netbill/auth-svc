package responses

import (
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/oapi"
)

func TokensPair(m models.TokensPair) oapi.TokensPair {
	resp := oapi.TokensPair{
		Data: oapi.TokensPairData{
			Id:   m.SessionID,
			Type: "tokens_pair",
			Attributes: oapi.TokensPairDataAttributes{
				AccessToken:  m.Access,
				RefreshToken: m.Refresh,
			},
		},
	}

	return resp
}
