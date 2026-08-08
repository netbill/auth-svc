package errx

import (
	"github.com/netbill/ape"
)

var (
	ErrorAccountNotFound = ape.DeclareError("ACCOUNT_NOT_FOUND")
	ErrorAccountDeleted  = ape.DeclareError("ACCOUNT_DELETED")

	ErrorAccountInvalidSession = ape.DeclareError("ACCOUNT_INVALID_SESSION")

	ErrorEmailAlreadyExist = ape.DeclareError("EMAIL_ALREADY_EXIST")

	ErrorPasswordInvalid         = ape.DeclareError("PASSWORD_INVALID")
	ErrorPasswordIsNotAllowed    = ape.DeclareError("PASSWORD_IS_NOT_ALLOWED")
	ErrorCannotChangePasswordYet = ape.DeclareError("CANNOT_CHANGE_PASSWORD_YET")

	ErrorRoleNotSupported = ape.DeclareError("ACCOUNT_ROLE_NOT_SUPPORTED")
)
