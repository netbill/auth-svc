package organization

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

type Module struct {
	repo repo
}

func New(repo repo) *Module {
	return &Module{repo: repo}
}

type repo interface {
	CreateOrgMember(ctx context.Context, member models.Member) error
	DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error
}
