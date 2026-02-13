package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

type AccountEmailRow struct {
	AccountID uuid.UUID `db:"account_id"`
	Email     string    `db:"email"`
	Verified  bool      `db:"verified"`
	UpdatedAt time.Time `db:"updated_at"`
	CreatedAt time.Time `db:"created_at"`
}

func (a AccountEmailRow) IsNil() bool {
	return a.AccountID == uuid.Nil
}

func (a AccountEmailRow) ToModel() models.AccountEmail {
	return models.AccountEmail{
		AccountID: a.AccountID,
		Email:     a.Email,
		Verified:  a.Verified,
		UpdatedAt: a.UpdatedAt,
		CreatedAt: a.CreatedAt,
	}
}

type AccountEmailsQ interface {
	New() AccountEmailsQ
	Insert(ctx context.Context, input AccountEmailRow) (AccountEmailRow, error)

	Get(ctx context.Context) (AccountEmailRow, error)
	Select(ctx context.Context) ([]AccountEmailRow, error)

	UpdateMany(ctx context.Context) (int64, error)
	UpdateOne(ctx context.Context) (AccountEmailRow, error)

	UpdateEmail(email string) AccountEmailsQ
	UpdateVerified(verified bool) AccountEmailsQ

	Delete(ctx context.Context) error

	FilterAccountID(accountID uuid.UUID) AccountEmailsQ
	FilterEmail(email string) AccountEmailsQ

	Exists(ctx context.Context) (bool, error)
}

func (r *Repository) ExistsAccountByEmail(ctx context.Context, email string) (bool, error) {
	exist, err := r.AccountEmailsQ.FilterEmail(email).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence by email %s, cause: %w", email, err)
	}

	return exist, nil
}

func (r *Repository) GetAccountEmail(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error) {
	acc, err := r.AccountEmailsQ.FilterAccountID(accountID).Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.AccountEmail{}, errx.ErrorAccountEmailNotFound.Raise(err)
	case err != nil:
		return models.AccountEmail{}, fmt.Errorf(
			"failed to get account email for account %s, cause: %w", accountID, err,
		)
	}

	return acc.ToModel(), nil
}
