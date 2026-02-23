package handler

import (
	"context"
	"encoding/json"

	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/eventbox"
	"github.com/netbill/evtypes"
)

func (h *Handler) OrgMemberCreated(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrgMemberCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	return h.modules.org.CreateOrgMember(ctx, models.OrgMember{
		ID:             payload.MemberID,
		AccountID:      payload.AccountID,
		OrganizationID: payload.OrganizationID,
		CreatedAt:      payload.CreatedAt,
	})
}

func (h *Handler) OrgMemberDeleted(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrgMemberDeletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	return h.modules.org.DeleteOrgMember(ctx, payload.MemberID)
}
