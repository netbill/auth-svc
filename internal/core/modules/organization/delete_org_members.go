package organization

import (
	"context"

	"github.com/google/uuid"
)

func (m *Module) DeleteOrgMembers(ctx context.Context, orgID uuid.UUID) error {
	return m.repo.DeleteOrgMembers(ctx, orgID)
}
