package errx

import (
	"github.com/netbill/ape"
)

var ErrorAccountNotFound = ape.DeclareError("ACCOUNT_NOT_FOUND")

var ErrorVersionConflict = ape.DeclareError("VERSION_CONFLICT")

var ErrorUsernameAlreadyTaken = ape.DeclareError("USERNAME_ALREADY_TAKEN")
var ErrorUsernameIsNotAllowed = ape.DeclareError("USERNAME_IS_NOT_ALLOWED")

var ErrorInitiatorNotFound = ape.DeclareError("INITIATOR_NOT_FOUND")
var ErrorInitiatorInvalidSession = ape.DeclareError("INITIATOR_INVALID_SESSION")

var ErrorAccountEmailNotFound = ape.DeclareError("ACCOUNT_EMAIL_NOT_FOUND")

var ErrorEmailAlreadyExist = ape.DeclareError("EMAIL_ALREADY_EXIST")
var ErrorEmailNotVerified = ape.DeclareError("EMAIL_NOT_VERIFIED")

var ErrorAccountPasswordNorFound = ape.DeclareError("ACCOUNT_PASSWORD_NOR_FOUND")

var ErrorPasswordInvalid = ape.DeclareError("PASSWORD_INVALID")
var ErrorPasswordIsNotAllowed = ape.DeclareError("PASSWORD_IS_NOT_ALLOWED")

var ErrorCannotChangePasswordYet = ape.DeclareError("CANNOT_CHANGE_PASSWORD_YET")

var ErrorRoleNotSupported = ape.DeclareError("ACCOUNT_ROLE_NOT_SUPPORTED")
var AccountHaveMembershipInOrg = ape.DeclareError("CANNOT_DELETE_ACCOUNT_ORG_MEMBER")
