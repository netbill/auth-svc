package evcontroller

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/eventbox"
	"github.com/netbill/evtypes"
)

const operationOrgMemberCreated = "organization_member_created"

func (c *OrgController) OrgMemberCreated(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrgMemberCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	log := c.log.WithOperation(operationOrgMemberCreated).
		With(slog.String("member_id", payload.MemberID.String()))

	err := c.modules.org.CreateMember(ctx, models.OrgMember{
		ID:             payload.MemberID,
		AccountID:      payload.AccountID,
		OrganizationID: payload.OrganizationID,
		CreatedAt:      payload.CreatedAt,
	})
	switch {
	case errors.Is(err, errx.ErrorOrgMemberDeleted):
		log.Debug("received org member created already deleted org member")
		return nil
	case errors.Is(err, errx.ErrorAccountDeleted):
		log.Debug("received org member created event for already deleted account")
		return nil
	case errors.Is(err, errx.ErrorOrganizationDeleted):
		log.Debug("received org member created event for already deleted organization")
		return nil
	case errors.Is(err, errx.ErrorOrgMemberAlreadyExists):
		log.Debug("received org member created event for already existing org member")
		return nil
	case err != nil:
		log.WithError(err).Error("failed to create org member")
		return err
	default:
		log.Debug("org member created successfully")
		return nil
	}
}

const operationOrgMemberDeleted = "organization_member_deleted"

func (c *OrgController) OrgMemberDeleted(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrgMemberDeletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	log := c.log.WithOperation(operationOrgMemberDeleted).
		With(slog.String("member_id", payload.MemberID.String()))

	err := c.modules.org.DeleteMember(ctx, payload.MemberID)
	switch {
	case errors.Is(err, errx.ErrorOrgMemberDeleted):
		log.Debug("received org member deleted event for already deleted org member")
		return nil
	case errors.Is(err, errx.ErrorAccountDeleted):
		log.Debug("received org member deleted event for already deleted account")
		return nil
	case errors.Is(err, errx.ErrorOrganizationDeleted):
		log.Debug("received org member deleted event for already deleted organization")
		return nil
	case err != nil:
		log.WithError(err).Error("failed to delete org member")
		return err
	default:
		log.Debug("org member deleted successfully")
		return nil
	}
}
