package errx

import (
	"github.com/netbill/ape"
)

var ErrorAccountNotFound = ape.DeclareError("ACCOUNT_NOT_FOUND")

var ErrorAccountInvalidSession = ape.DeclareError("ACCOUNT_INVALID_SESSION")

var ErrorUsernameAlreadyTaken = ape.DeclareError("USERNAME_ALREADY_TAKEN")
var ErrorUsernameIsNotAllowed = ape.DeclareError("USERNAME_IS_NOT_ALLOWED")
var ErrorEmailAlreadyExist = ape.DeclareError("EMAIL_ALREADY_EXIST")

var ErrorPasswordInvalid = ape.DeclareError("PASSWORD_INVALID")
var ErrorPasswordIsNotAllowed = ape.DeclareError("PASSWORD_IS_NOT_ALLOWED")
var ErrorCannotChangePasswordYet = ape.DeclareError("CANNOT_CHANGE_PASSWORD_YET")

var ErrorRoleNotSupported = ape.DeclareError("ACCOUNT_ROLE_NOT_SUPPORTED")
var AccountHaveMembershipInOrg = ape.DeclareError("CANNOT_DELETE_ACCOUNT_ORG_MEMBER")
