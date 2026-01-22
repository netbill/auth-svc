package organization

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
)

func (m Module) DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error {
	err := m.repo.DeleteOrgMember(ctx, memberID)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("delete org member error: %w", err),
		)
	}

	return nil
}
