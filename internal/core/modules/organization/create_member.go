package organization

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m Module) CreateOrgMember(ctx context.Context, member models.Member) error {
	err := m.repo.CreateOrgMember(ctx, member)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("create org member error: %w", err),
		)
	}

	return nil
}
