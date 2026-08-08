package username

import (
	"fmt"
	"math/rand/v2"
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

var adjectives = []string{
	"Swift", "Brave", "Calm", "Eager", "Fuzzy", "Gentle", "Happy", "Jolly",
	"Keen", "Lively", "Mighty", "Nimble", "Proud", "Quiet", "Rapid", "Sharp",
	"Tidy", "Vivid", "Witty", "Zesty",
}

var nouns = []string{
	"Falcon", "Otter", "Tiger", "Panda", "Eagle", "Wolf", "Fox", "Bear",
	"Hawk", "Lynx", "Raven", "Shark", "Whale", "Heron", "Cobra", "Puma",
	"Ibex", "Moose", "Crane", "Viper",
}

func (v *Validator) Generate() string {
	adjective := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]
	suffix := rand.IntN(1_000_000)

	return fmt.Sprintf("%s%s%06d", adjective, noun, suffix)
}
