package responses

import (
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/oapi"
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
	m models.User,
	opts ...UserOption,
) oapi.User {
	res := &userResponse{
		user: m,
	}
	for _, opt := range opts {
		opt(res)
	}

	data := oapi.UserData{
		Id:   m.ID,
		Type: "user",
		Attributes: oapi.UserDataAttributes{
			Role:      m.Role,
			Version:   m.Version,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
	}

	included := make([]oapi.UserEmail, 0)
	if res.email != nil {
		included = append(included, UserEmailData(*res.email))
	}

	return oapi.User{
		Data:     data,
		Included: included,
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
