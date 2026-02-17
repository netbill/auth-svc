package responses

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/core/models"
	resources2 "github.com/netbill/auth-svc/pkg/resources"
	"github.com/netbill/restkit/pagi"
)

func AccountSession(m models.Session) resources2.AccountSession {
	resp := resources2.AccountSession{
		Data: resources2.AccountSessionData{
			Id:   m.ID,
			Type: "account_session",
			Attributes: resources2.AccountSessionAttributes{
				AccountId: m.AccountID,
				Version:   m.Version,
				CreatedAt: m.CreatedAt,
				LastUsed:  m.LastUsed,
			},
		},
	}

	return resp
}

func AccountSessionsCollection(r *http.Request, page pagi.Page[[]models.Session]) resources2.AccountSessionsCollection {
	data := make([]resources2.AccountSessionData, 0, len(page.Data))

	for _, s := range page.Data {
		data = append(data, AccountSession(s).Data)
	}

	links := pagi.BuildPageLinks(r, page.Page, page.Size, page.Total)

	return resources2.AccountSessionsCollection{
		Data: data,
		Links: resources2.PaginationData{
			First: links.First,
			Last:  links.Last,
			Prev:  links.Prev,
			Next:  links.Next,
			Self:  links.Self,
		},
	}
}
