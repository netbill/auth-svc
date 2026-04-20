package passmanager

import (
	"errors"
	"fmt"

	"github.com/netbill/auth-svc/internal/errx"
	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	cost int
}

// New - create new password manager
// bcrypt.MinCost     int = 4  // the minimum allowable cost as passed in to bcrypt.GenerateFromPassword
// bcrypt.MaxCost     int = 31 // the maximum allowable cost as passed in to bcrypt.GenerateFromPassword
// bcrypt.DefaultCost int = 10 // the cost that will actually be set if a cost below MinCost is passed into bcrypt.GenerateFromPassword
func New(cost int) *Manager {
	if cost < 4 || cost > 31 {
		panic("bcrypt cost must be between 4 and 31")
	}

	return &Manager{
		cost: cost,
	}
}

func (m *Manager) GenerateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), m.cost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash, cause: %w", err)
	}

	return string(hash), nil
}
func (m *Manager) CheckMatch(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errx.ErrorPasswordInvalid.Raise(
				fmt.Errorf("invalid credentials, cause: %w", err),
			)
		}

		return fmt.Errorf("comparing password hash, cause: %w", err)
	}

	return nil
}
