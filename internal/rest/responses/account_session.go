package responses

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/resources"
	"github.com/netbill/restkit/pagi"
)

func AccountSession(m models.Session) resources.AccountSession {
	resp := resources.AccountSession{
		Data: resources.AccountSessionData{
			Id:   m.ID,
			Type: "account_session",
			Attributes: resources.AccountSessionAttributes{
				AccountId: m.AccountID,
				Version:   m.Version,
				CreatedAt: m.CreatedAt,
				LastUsed:  m.LastUsed,
			},
		},
	}

	return resp
}

func AccountSessionsCollection(r *http.Request, page pagi.Page[[]models.Session]) resources.AccountSessionsCollection {
	data := make([]resources.AccountSessionData, 0, len(page.Data))

	for _, s := range page.Data {
		data = append(data, AccountSession(s).Data)
	}

	links := pagi.BuildPageLinks(r, page.Page, page.Size, page.Total)

	return resources.AccountSessionsCollection{
		Data: data,
		Links: resources.PaginationData{
			First: links.First,
			Last:  links.Last,
			Prev:  links.Prev,
			Next:  links.Next,
			Self:  links.Self,
		},
	}
}
