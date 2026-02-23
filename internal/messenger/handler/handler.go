package handler

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

type Handler struct {
	modules
}

type modules struct {
	org orgModule
}

func New(org orgModule) *Handler {
	return &Handler{
		modules: modules{
			org: org,
		},
	}
}

type orgModule interface {
	CreateOrgMember(ctx context.Context, member models.OrgMember) error
	DeleteOrgMembers(ctx context.Context, orgID uuid.UUID) error
	DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error
}
