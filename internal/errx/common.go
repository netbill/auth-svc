package errx

import (
	"errors"

	"github.com/netbill/ape"
)

var ErrorForbidden = ape.DeclareError("FORBIDDEN")

var ErrorNotEnoughRights = ape.DeclareError("NOT_ENOUGH_RIGHTS")

var ErrCacheMiss = errors.New("cache miss")
