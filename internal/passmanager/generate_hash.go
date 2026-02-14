package passmanager

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func (p *Passer) GenerateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash, cause: %w", err)
	}

	return string(hash), nil
}
