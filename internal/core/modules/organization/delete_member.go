package organization

import (
	"context"

	"github.com/google/uuid"
)

func (m *Module) DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error {
	err := m.repo.DeleteOrgMember(ctx, memberID)
	if err != nil {
		return err
	}

	return nil
}
