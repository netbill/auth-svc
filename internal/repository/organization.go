package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
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

type OrganizationRepo struct {
	query OrganizationsQ
}

func NewOrganizationRepo(query OrganizationsQ) *OrganizationRepo {
	return &OrganizationRepo{
		query: query,
	}
}

func (r *OrganizationRepo) Create(ctx context.Context, organization models.Organization) error {
	return r.query.New().Insert(ctx, OrganizationRow{
		ID:              organization.ID,
		SourceCreatedAt: organization.CreatedAt,
	})
}

func (r *OrganizationRepo) Delete(ctx context.Context, organizationID uuid.UUID) error {
	return r.query.New().FilterByID(organizationID).Delete(ctx)
}

func (r *OrganizationRepo) Get(ctx context.Context, organizationID uuid.UUID) (models.Organization, error) {
	row, err := r.query.New().FilterByID(organizationID).Get(ctx)
	if err != nil {
		return models.Organization{}, err
	}
	if row.IsNil() {
		return models.Organization{}, nil
	}

	return row.ToModel(), nil
}
