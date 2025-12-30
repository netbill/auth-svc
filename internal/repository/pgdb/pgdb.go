package pgdb

import (
	"github.com/umisto/sso-svc/internal/domain/models"
)

func (a *Account) ToModel() models.Account {
	return models.Account{
		ID:                a.ID,
		Username:          a.Username,
		Role:              a.Role,
		Status:            a.Status,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
		UsernameUpdatedAt: a.UsernameUpdatedAt,
	}
}

func (a *AccountPassword) ToModel() models.AccountPassword {
	return models.AccountPassword{
		AccountID: a.AccountID,
		Hash:      a.Hash,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func (e *AccountEmail) ToModel() models.AccountEmail {
	return models.AccountEmail{
		AccountID: e.AccountID,
		Email:     e.Email,
		Verified:  e.Verified,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func (s *Session) ToModel() models.Session {
	return models.Session{
		ID:        s.ID,
		AccountID: s.AccountID,
		LastUsed:  s.LastUsed,
		CreatedAt: s.CreatedAt,
	}
}
