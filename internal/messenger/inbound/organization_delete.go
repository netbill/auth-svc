package inbound

import (
	"context"
	"encoding/json"

	"github.com/netbill/auth-svc/internal/messenger/evtypes"
	"github.com/segmentio/kafka-go"
)

func (i *Inbound) OrgDeleted(
	ctx context.Context,
	message kafka.Message,
) error {
	var payload evtypes.OrgDeletedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return err
	}

	return i.modules.org.DeleteOrgMembers(ctx, payload.OrganizationID)
}
