package evtypes

import (
	"time"

	"github.com/google/uuid"
)

const OrgMemberTopicV1 = "organization.member.v1"

const OrgMemberCreatedEvent = "member.created"

type OrgMemberCreatedPayload struct {
	MemberID       uuid.UUID `json:"member_id"`
	AccountID      uuid.UUID `json:"account_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Position       *string   `json:"position,omitempty"`
	Label          *string   `json:"label,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

const OrgMemberUpdatedEvent = "member.updated"

type OrgMemberUpdatedPayload struct {
	MemberID  uuid.UUID `json:"member_id"`
	Position  *string   `json:"position,omitempty"`
	Label     *string   `json:"label,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

const OrgMemberDeletedEvent = "member.deleted"

type OrgMemberDeletedPayload struct {
	MemberID  uuid.UUID `json:"member_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

const OrgTopicV1 = "organization.v1"

const OrgDeletedEvent = "organization.deleted"

type OrgDeletedPayload struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	DeletedAt      time.Time `json:"deleted_at"`
}
