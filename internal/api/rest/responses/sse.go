package responses

import (
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/pkg/oapi"
)

// QRTokenEvent builds the body of the qr_token SSE event. token is always a
// UUID string minted by session.CreateQRToken.
func QRTokenEvent(token string) oapi.QRToken {
	return oapi.QRToken{
		Data: oapi.QRTokenData{
			Type: "qr_token",
			Attributes: oapi.QRTokenDataAttributes{
				QrToken: uuid.MustParse(token),
			},
		},
	}
}
