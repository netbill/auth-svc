package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/logium"
)

type Inbound struct {
	log    *logium.Logger
	domain domain
}

func New(log *logium.Logger, domain domain) *Inbound {
	return &Inbound{
		log:    log,
		domain: domain,
	}
}

type domain interface {
	CreateOrgMember(ctx context.Context, member models.Member) error
	DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error
}
