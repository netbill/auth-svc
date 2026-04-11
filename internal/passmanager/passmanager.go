package passmanager

import (
	"github.com/netbill/auth-svc/internal/errx"
	"golang.org/x/crypto/bcrypt"
)

const defaultCost = bcrypt.DefaultCost

type Manager struct{}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) GenerateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), defaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (m *Manager) CheckMatch(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return errx.ErrorPasswordInvalid.Raise(err)
	}
	return nil
}
