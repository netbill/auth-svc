package username

import (
	"fmt"
	"regexp"

	"github.com/netbill/auth-svc/internal/errx"
)

const (
	MinLength = 3
	MaxLength = 32
)

type Validator struct {
	reg *regexp.Regexp
}

func NewValidator() *Validator {
	return &Validator{
		reg: regexp.MustCompile("^[a-zA-Z0-9]+$"),
	}
}

func (v *Validator) Validate(username string) error {
	switch {
	case len(username) < MinLength || len(username) > MaxLength:
		return errx.ErrorUsernameNotValid.Raise(
			fmt.Errorf("username must be between %d and %d characters", MinLength, MaxLength),
		)
	case !v.reg.MatchString(username):
		return errx.ErrorUsernameNotValid.Raise(
			fmt.Errorf("username must match regex %q", v.reg),
		)
	}

	return nil
}
