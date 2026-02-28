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

func (h *Handler) OrgDeleted(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrganizationDeletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	log := h.log.WithInboxEvent(event)

	err := h.modules.org.Delete(ctx, payload.OrganizationID)
	switch {
	case errors.Is(err, errx.ErrorOrganizationDeleted):
		log.WithInboxEvent(event).Debug("received organization deleted event for already deleted organization")
		return nil
	case err != nil:
		return err
	default:
		log.Debug("organization deleted successfully")
		return nil
	}
}

func (h *Handler) OrgCreated(
	ctx context.Context,
	event eventbox.InboxEvent,
) error {
	var payload evtypes.OrganizationCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	log := h.log.WithInboxEvent(event)

	err := h.modules.org.Create(ctx, models.Organization{
		ID:        payload.OrganizationID,
		CreatedAt: payload.CreatedAt,
	})
	switch {
	case errors.Is(err, errx.ErrorOrganizationDeleted):
		log.Debug("received organization created event for already deleted organization")
		return nil
	case errors.Is(err, errx.ErrorOrganizationAlreadyExists):
		log.Debug("received organization created event for already existing organization")
		return nil
	case err != nil:
		return err
	default:
		log.Debug("organization created successfully")
		return nil
	}
}
