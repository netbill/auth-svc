package inbound

import (
	"context"
	"encoding/json"

	"github.com/netbill/auth-svc/internal/messenger/evtypes"
	"github.com/segmentio/kafka-go"
)

func (i *Inbound) OrgMemberDeleted(
	ctx context.Context,
	message kafka.Message,
) error {
	var payload evtypes.OrgMemberDeletedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return err
	}

	return i.modules.org.DeleteOrgMember(ctx, payload.MemberID)
}
