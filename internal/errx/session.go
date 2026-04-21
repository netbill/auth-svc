package errx

import (
	"github.com/netbill/ape"
)

var (
	ErrorSessionNotFound = ape.DeclareError("SESSION_NOT_FOUND")
	ErrorSessionDeleted  = ape.DeclareError("SESSION_DELETED")

	ErrorSessionTokenMismatch = ape.DeclareError("SESSION_TOKEN_MISMATCH")
	ErrorSessionExpired       = ape.DeclareError("SESSION_EXPIRED")

	ErrorQRTokenNotFound         = ape.DeclareError("QR_TOKEN_NOT_FOUND")
	ErrorQRTokenAlreadyConfirmed = ape.DeclareError("QR_TOKEN_ALREADY_CONFIRMED")
)
