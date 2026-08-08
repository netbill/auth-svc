package responses

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/oapi"
)

func UploadUserMediaLinks(r *http.Request, u models.User, links models.UploadUserMediaLinks) oapi.UploadUserMediaLinks {
	return oapi.UploadUserMediaLinks{
		Data: oapi.UploadUserMediaLinksData{
			Id:   u.ID,
			Type: "user_upload_links",
			Attributes: oapi.UploadUserMediaLinksDataAttributes{
				Avatar: oapi.UploadUserMediaLinksDataAttributesAvatar{
					Key:        links.Avatar.Key,
					UploadUrl:  links.Avatar.UploadURL,
					PreloadUrl: links.Avatar.PreloadUrl,
				},
			},
			Relationships: oapi.UploadUserMediaLinksDataRelationships{
				User: &oapi.UploadUserMediaLinksDataRelationshipsUser{
					Data: oapi.UploadUserMediaLinksDataRelationshipsUserData{
						Id:   u.ID,
						Type: "user",
					},
				},
			},
		},
		Included: []oapi.UserData{
			userData(r, u),
		},
	}
}
