package account

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

// VerifyEmail is a callback function to verify email addresses, which we should use in Kafka consumer
func (s Service) VerifyEmail(ctx context.Context, email string) (models.AccountEmail, error) {
	account, err := s.GetAccountByEmail(ctx, email)
	if err != nil {
		return models.AccountEmail{}, err
	}

	emailData, err := s.repo.UpdateAccountEmailVerification(ctx, account.ID, true)
	if err != nil {
		return models.AccountEmail{}, errx.ErrorInternal.Raise(
			fmt.Errorf("verifying email for account %s, cause: %w", account.ID, err),
		)
	}

	return emailData, nil
}
