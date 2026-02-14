package passmanager

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/netbill/auth-svc/internal/core/errx"
)

func (p *Passer) CheckRequirements(password string) error {
	if len(password) < 8 || len(password) > 32 {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("password must be between 8 and 32 characters"),
		)
	}

	var (
		hasUpper, hasLower, hasDigit, hasSpecial bool
	)

	allowedSpecials := "-.!#$%&?,@"

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune(allowedSpecials, r):
			hasSpecial = true
		default:
			return errx.ErrorPasswordIsNotAllowed.Raise(
				fmt.Errorf("password contains invalid characters %s", string(r)),
			)
		}
	}

	if !hasUpper {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one uppercase letter"),
		)
	}
	if !hasLower {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one lower case letter"),
		)
	}
	if !hasDigit {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one digit"),
		)
	}
	if !hasSpecial {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one special character from %s", allowedSpecials),
		)
	}

	return nil
}
