package evcontroller

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/log"
)

type OrgController struct {
	log     *log.Logger
	modules modules
}

type modules struct {
	org orgCore
}

func New(log *log.Logger, org orgCore) *OrgController {
	return &OrgController{
		log: log,
		modules: modules{
			org: org,
		},
	}
}

type orgCore interface {
	Create(ctx context.Context, organization models.Organization) error
	Get(ctx context.Context, organizationID uuid.UUID) (models.Organization, error)
	Delete(ctx context.Context, organizationID uuid.UUID) error

	CreateMember(ctx context.Context, member models.OrgMember) error
	DeleteMember(ctx context.Context, memberID uuid.UUID) error
}
