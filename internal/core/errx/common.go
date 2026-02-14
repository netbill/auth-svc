package errx

import (
	"github.com/netbill/ape"
)

var ErrorForbidden = ape.DeclareError("FORBIDDEN")

var ErrorNotEnoughRights = ape.DeclareError("NOT_ENOUGH_RIGHTS")
