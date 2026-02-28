package messenger

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/eventbox"
	"github.com/netbill/evtypes"
)

type handlers interface {
	OrgMemberCreated(
		ctx context.Context,
		event eventbox.InboxEvent,
	) error
	OrgMemberDeleted(
		ctx context.Context,
		event eventbox.InboxEvent,
	) error
	OrgCreated(
		ctx context.Context,
		event eventbox.InboxEvent,
	) error
	OrgDeleted(
		ctx context.Context,
		event eventbox.InboxEvent,
	) error
}

func NewInboxWorker(
	logger *log.Logger,
	inbox eventbox.Inbox,
	cfg eventbox.InboxWorkerConfig,
	handlers handlers,
) *eventbox.InboxWorker {
	id := uuid.New().String()

	worker := eventbox.NewInboxWorker(id, logger, inbox, cfg)

	worker.Route(evtypes.OrganizationCreatedEvent, handlers.OrgCreated)
	worker.Route(evtypes.OrganizationDeletedEvent, handlers.OrgDeleted)

	worker.Route(evtypes.OrgMemberCreatedEvent, handlers.OrgMemberCreated)
	worker.Route(evtypes.OrgMemberDeletedEvent, handlers.OrgMemberDeleted)

	return worker
}
