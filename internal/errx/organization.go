package errx

import "github.com/netbill/ape"

var (
	ErrorOrgMemberNotFound      = ape.DeclareError("ORG_MEMBER_NOT_FOUND")
	ErrorOrgMemberAlreadyExists = ape.DeclareError("ORG_MEMBER_ALREADY_EXISTS")
	ErrorOrgMemberDeleted       = ape.DeclareError("ORG_MEMBER_DELETED")

	ErrorOrganizationNotFound      = ape.DeclareError("ORGANIZATION_NOT_FOUND")
	ErrorOrganizationAlreadyExists = ape.DeclareError("ORGANIZATION_ALREADY_EXISTS")
	ErrorOrganizationDeleted       = ape.DeclareError("ORGANIZATION_DELETED")
)
