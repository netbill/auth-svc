package responses

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/oapi"
	"github.com/netbill/restkit/pagi"
)

func AccountSession(m models.Session) oapi.AccountSession {
	attrs := oapi.AccountSessionAttributes{
		AccountId: m.AccountID,
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		LastUsed:  m.LastUsed,
		DeletedAt: m.DeletedAt,
	}

	return oapi.AccountSession{
		Data: oapi.AccountSessionData{
			Id:         m.ID,
			Type:       "account_session",
			Attributes: attrs,
		},
	}
}

func AccountSessionsCollection(r *http.Request, page pagi.Page[[]models.Session]) oapi.AccountSessionsCollection {
	data := make([]oapi.AccountSessionData, 0, len(page.Data))

	for _, s := range page.Data {
		data = append(data, AccountSession(s).Data)
	}

	links := pagi.BuildPageLinks(r, page.Page, page.Size, page.Total)

	return oapi.AccountSessionsCollection{
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
