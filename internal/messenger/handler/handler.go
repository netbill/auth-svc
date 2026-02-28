package handler

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/pkg/log"
)

type Handler struct {
	log     *log.Logger
	modules modules
}

type modules struct {
	org orgModule
}

func New(log *log.Logger, org orgModule) *Handler {
	return &Handler{
		log: log,
		modules: modules{
			org: org,
		},
	}
}

type orgModule interface {
	Create(ctx context.Context, organization models.Organization) error
	Get(ctx context.Context, orgID uuid.UUID) (models.Organization, error)
	Delete(ctx context.Context, orgID uuid.UUID) error

	CreateOrgMember(ctx context.Context, member models.OrgMember) error
	DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error
}
