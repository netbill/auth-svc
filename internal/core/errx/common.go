package errx

import (
	"github.com/netbill/restkit/ape"
)

var ErrorInternal = ape.DeclareError("INTERNAL_ERROR")

var ErrorForbidden = ape.DeclareError("FORBIDDEN")

var ErrorNotEnoughRights = ape.DeclareError("NOT_ENOUGH_RIGHTS")
