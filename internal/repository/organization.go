package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

type OrganizationRow struct {
	ID               uuid.UUID `db:"id"`
	SourceCreatedAt  time.Time `db:"source_created_at"`
	ReplicaCreatedAt time.Time `db:"replica_created_at"`
}

func (o OrganizationRow) IsNil() bool {
	return o.ID == uuid.Nil
}

func (o OrganizationRow) ToModel() models.Organization {
	return models.Organization{
		ID:        o.ID,
		CreatedAt: o.SourceCreatedAt,
	}
}

type OrganizationsQ interface {
	New() OrganizationsQ

	Insert(ctx context.Context, organization OrganizationRow) error
	FilterByID(id uuid.UUID) OrganizationsQ

	Get(ctx context.Context) (OrganizationRow, error)
	Select(ctx context.Context) ([]OrganizationRow, error)

	Exists(ctx context.Context) (bool, error)
	Delete(ctx context.Context) error
}

func (r *Repository) CreateOrganization(ctx context.Context, organization models.Organization) error {
	return r.OrganizationsSql.New().Insert(ctx, OrganizationRow{
		ID:              organization.ID,
		SourceCreatedAt: organization.CreatedAt,
	})
}

func (r *Repository) DeleteOrganization(ctx context.Context, orgID uuid.UUID) error {
	return r.OrganizationsSql.New().FilterByID(orgID).Delete(ctx)
}

func (r *Repository) GetOrganizationByID(ctx context.Context, orgID uuid.UUID) (models.Organization, error) {
	row, err := r.OrganizationsSql.New().FilterByID(orgID).Get(ctx)
	if err != nil {
		return models.Organization{}, err
	}
	if row.IsNil() {
		return models.Organization{}, nil
	}

	return row.ToModel(), nil
}
