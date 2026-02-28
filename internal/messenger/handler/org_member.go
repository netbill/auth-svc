package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/netbill/auth-svc/internal/core/errx"
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

	log := h.log.WithInboxEvent(event)

	err := h.modules.org.CreateOrgMember(ctx, models.OrgMember{
		ID:             payload.MemberID,
		AccountID:      payload.AccountID,
		OrganizationID: payload.OrganizationID,
		CreatedAt:      payload.CreatedAt,
	})
	switch {
	case errors.Is(err, errx.ErrorOrgMemberDeleted):
		log.Debug("received org member created event for already deleted org member")
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
		return err
	default:
		log.Debug("org member created successfully")
		return nil
	}
}

func (h *Handler) OrgMemberDeleted(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrgMemberDeletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	err := h.modules.org.DeleteOrgMember(ctx, payload.MemberID)
	switch {
	case errors.Is(err, errx.ErrorOrgMemberDeleted):
		h.log.WithInboxEvent(event).Debug("received org member deleted event for already deleted org member")
		return nil
	case errors.Is(err, errx.ErrorAccountDeleted):
		h.log.WithInboxEvent(event).Debug("received org member deleted event for already deleted account")
		return nil
	case errors.Is(err, errx.ErrorOrganizationDeleted):
		h.log.WithInboxEvent(event).Debug("received org member deleted event for already deleted organization")
		return nil
	case err != nil:
		return err
	default:
		h.log.WithInboxEvent(event).Debug("org member deleted successfully")
		return nil
	}
}
