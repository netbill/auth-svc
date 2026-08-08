package responses

import (
	"net/http"

	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/oapi"
	"github.com/netbill/restkit/pagi"
)

type userResponse struct {
	user  models.User
	email *models.UserEmail
}

type UserOption func(*userResponse)

func WithUserEmail(email models.UserEmail) UserOption {
	return func(res *userResponse) {
		res.email = &email
	}
}

func User(
	r *http.Request,
	m models.User,
	opts ...UserOption,
) oapi.User {
	res := &userResponse{
		user: m,
	}
	for _, opt := range opts {
		opt(res)
	}

	included := make([]oapi.UserEmail, 0)
	if res.email != nil {
		included = append(included, UserEmailData(*res.email))
	}

	return oapi.User{
		Data:     userData(r, m),
		Included: included,
	}
}

func userData(r *http.Request, m models.User) oapi.UserData {
	res := oapi.UserData{
		Id:   m.ID,
		Type: "user",
		Attributes: oapi.UserDataAttributes{
			Username:    m.Username,
			Pseudonym:   m.Pseudonym,
			Description: m.Description,
			Role:        m.Role,
			Version:     m.Version,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		},
	}
	if m.AvatarKey != nil {
		url := scope.ResolverURL(r, *m.AvatarKey)
		res.Attributes.AvatarUrl = &url
	}

	return res
}

func UserCollection(r *http.Request, page pagi.Page[[]models.User]) oapi.UsersCollection {
	data := make([]oapi.UserData, len(page.Data))
	for i, u := range page.Data {
		data[i] = userData(r, u)
	}

	links := pagi.BuildPageLinks(r, page.Page, page.Size, page.Total)

	return oapi.UsersCollection{
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

func UserEmailData(ae models.UserEmail) oapi.UserEmail {
	return oapi.UserEmail{
		Id:   ae.UserID,
		Type: "user_email",
		Attributes: oapi.UserEmailAttributes{
			Email:     ae.Email,
			Version:   ae.Version,
			Verified:  ae.Verified,
			UpdatedAt: ae.UpdatedAt,
		},
	}
}
