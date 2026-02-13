package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

type Inbound struct {
	modules
}

type modules struct {
	org orgModule
}

func New(org orgModule) *Inbound {
	return &Inbound{
		modules: modules{
			org: org,
		},
	}
}

type orgModule interface {
	CreateOrgMember(ctx context.Context, member models.OrgMember) error
	DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error
}
