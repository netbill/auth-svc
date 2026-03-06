package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	errx2 "github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/eventbox"
	"github.com/netbill/evtypes"
)

const operationOrgMemberCreated = "organization_member_created"

func (h *Handler) OrgMemberCreated(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrgMemberCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	log := h.log.WithOperation(operationOrgMemberCreated).
		With(slog.String("member_id", payload.MemberID.String()))

	err := h.modules.org.CreateMember(ctx, models.OrgMember{
		ID:             payload.MemberID,
		AccountID:      payload.AccountID,
		OrganizationID: payload.OrganizationID,
		CreatedAt:      payload.CreatedAt,
	})
	switch {
	case errors.Is(err, errx2.ErrorOrgMemberDeleted):
		log.Debug("received org member created already deleted org member")
		return nil
	case errors.Is(err, errx2.ErrorAccountDeleted):
		log.Debug("received org member created event for already deleted account")
		return nil
	case errors.Is(err, errx2.ErrorOrganizationDeleted):
		log.Debug("received org member created event for already deleted organization")
		return nil
	case errors.Is(err, errx2.ErrorOrgMemberAlreadyExists):
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

func (h *Handler) OrgMemberDeleted(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrgMemberDeletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	log := h.log.WithOperation(operationOrgMemberDeleted).
		With(slog.String("member_id", payload.MemberID.String()))

	err := h.modules.org.DeleteMember(ctx, payload.MemberID)
	switch {
	case errors.Is(err, errx2.ErrorOrgMemberDeleted):
		log.Debug("received org member deleted event for already deleted org member")
		return nil
	case errors.Is(err, errx2.ErrorAccountDeleted):
		log.Debug("received org member deleted event for already deleted account")
		return nil
	case errors.Is(err, errx2.ErrorOrganizationDeleted):
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
