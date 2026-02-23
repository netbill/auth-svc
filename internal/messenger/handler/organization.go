package handler

import (
	"context"
	"encoding/json"

	"github.com/netbill/eventbox"
	"github.com/netbill/evtypes"
)

func (h *Handler) OrgDeleted(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrganizationDeletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	return h.modules.org.DeleteOrgMembers(ctx, payload.OrganizationID)
}
