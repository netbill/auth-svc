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

type AccountPasswordRow struct {
	AccountID uuid.UUID `db:"account_id"`
	Hash      string    `db:"hash"`
	UpdatedAt time.Time `db:"updated_at"`
	CreatedAt time.Time `db:"created_at"`
}

func (a AccountPasswordRow) IsNil() bool {
	return a.AccountID == uuid.Nil
}

func (a AccountPasswordRow) ToModel() models.AccountPassword {
	return models.AccountPassword{
		AccountID: a.AccountID,
		Hash:      a.Hash,
		UpdatedAt: a.UpdatedAt,
		CreatedAt: a.CreatedAt,
	}
}

type AccountPasswordsQ interface {
	New() AccountPasswordsQ
	Insert(ctx context.Context, input AccountPasswordRow) (AccountPasswordRow, error)

	Get(ctx context.Context) (AccountPasswordRow, error)
	Select(ctx context.Context) ([]AccountPasswordRow, error)

	UpdateMany(ctx context.Context) (int64, error)
	UpdateOne(ctx context.Context) (AccountPasswordRow, error)

	UpdateHash(hash string) AccountPasswordsQ

	Delete(ctx context.Context) error

	FilterAccountID(accountID uuid.UUID) AccountPasswordsQ

	Exists(ctx context.Context) (bool, error)
}

func (r *Repository) GetAccountPassword(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error) {
	acc, err := r.accountPasswordsQ().FilterAccountID(accountID).Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.AccountPassword{}, errx.ErrorAccountPasswordNorFound.Raise(
			fmt.Errorf("account password for account %s not found", accountID),
		)
	case err != nil:
		return models.AccountPassword{}, fmt.Errorf(
			"failed to get account password for account %s, cause: %w", accountID, err,
		)
	}

	return acc.ToModel(), nil
}

func (r *Repository) UpdateAccountPassword(
	ctx context.Context,
	accountID uuid.UUID,
	passwordHash string,
) (models.AccountPassword, error) {
	acc, err := r.accountPasswordsQ().
		FilterAccountID(accountID).
		UpdateHash(passwordHash).
		UpdateOne(ctx)
	if err != nil {
		return models.AccountPassword{}, fmt.Errorf(
			"failed to update account password for account %s, cause: %w", accountID, err,
		)
	}

	return acc.ToModel(), nil
}
