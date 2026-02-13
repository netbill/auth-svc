package organization

import (
	"context"

	"github.com/google/uuid"
)

func (m *Module) DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error {
	return m.repo.DeleteOrgMember(ctx, memberID)
}
