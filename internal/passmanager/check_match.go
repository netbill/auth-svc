package passmanager

import (
	"errors"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"golang.org/x/crypto/bcrypt"
)

func (p *Passer) CheckPasswordMatch(hash, password string) error {
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
