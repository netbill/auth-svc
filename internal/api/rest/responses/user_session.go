package responses

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/oapi"
	"github.com/netbill/restkit/pagi"
)

func UserSession(m models.Session) oapi.UserSession {
	attrs := oapi.UserSessionAttributes{
		UserId:    m.UserID,
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		LastUsed:  m.LastUsed,
		DeletedAt: m.DeletedAt,
	}

	return oapi.UserSession{
		Data: oapi.UserSessionData{
			Id:         m.ID,
			Type:       "user_session",
			Attributes: attrs,
		},
	}
}

func UserSessionsCollection(r *http.Request, page pagi.Page[[]models.Session]) oapi.UserSessionsCollection {
	data := make([]oapi.UserSessionData, 0, len(page.Data))

	for _, s := range page.Data {
		data = append(data, UserSession(s).Data)
	}

	links := pagi.BuildPageLinks(r, page.Page, page.Size, page.Total)

	return oapi.UserSessionsCollection{
		Data: data,
		Links: oapi.PaginationData{
			First: links.First,
			Last:  links.Last,
			Prev:  links.Prev,
			Next:  links.Next,
			Self:  links.Self,
		},
	}
}
