package inbound

import (
	"context"
	"encoding/json"

	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/messenger/evtypes"
	"github.com/segmentio/kafka-go"
)

func (i *Inbound) OrgMemberCreated(
	ctx context.Context,
	message kafka.Message,
) error {
	var payload evtypes.OrgMemberCreatedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return err
	}

	return i.modules.org.CreateOrgMember(ctx, models.OrgMember{
		ID:             payload.MemberID,
		AccountID:      payload.AccountID,
		OrganizationID: payload.OrganizationID,
		CreatedAt:      payload.CreatedAt,
	})
}
