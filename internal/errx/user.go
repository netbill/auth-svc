package errx

import (
	"github.com/netbill/ape"
)

var (
	ErrorUserNotFound = ape.DeclareError("USER_NOT_FOUND")
	ErrorUserDeleted  = ape.DeclareError("USER_DELETED")

	ErrorUserInvalidSession = ape.DeclareError("USER_INVALID_SESSION")

	ErrorEmailAlreadyExist = ape.DeclareError("EMAIL_ALREADY_EXIST")

	ErrorPasswordInvalid         = ape.DeclareError("PASSWORD_INVALID")
	ErrorPasswordIsNotAllowed    = ape.DeclareError("PASSWORD_IS_NOT_ALLOWED")
	ErrorCannotChangePasswordYet = ape.DeclareError("CANNOT_CHANGE_PASSWORD_YET")

	ErrorRoleNotSupported = ape.DeclareError("USER_ROLE_NOT_SUPPORTED")

	ErrorUserUploadedAvatarInvalid = ape.DeclareError("USER_UPLOADED_AVATAR_INVALID")
	ErrorUsernameNotValid          = ape.DeclareError("USERNAME_NOT_VALID")
	ErrorUsernameTaken             = ape.DeclareError("USERNAME_TAKEN")
)
