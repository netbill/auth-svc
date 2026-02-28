package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

type AccountPasswordRow struct {
	AccountID uuid.UUID `db:"account_id"`
	Hash      string    `db:"hash"`
	Version   int32     `db:"version"`
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
		Version:   a.Version,
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
	row, err := r.AccountPassSql.New().FilterAccountID(accountID).Get(ctx)
	switch {
	case err != nil:
		return models.AccountPassword{}, fmt.Errorf(
			"failed to get account password for account %s, cause: %w", accountID, err,
		)
	case row.IsNil():
		return models.AccountPassword{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account password for account %s not found", accountID),
		)
	}

	return row.ToModel(), nil
}

func (r *Repository) UpdateAccountPassword(
	ctx context.Context,
	accountID uuid.UUID,
	passwordHash string,
) (models.AccountPassword, error) {
	acc, err := r.AccountPassSql.New().
		FilterAccountID(accountID).
		UpdateHash(passwordHash).
		UpdateOne(ctx)
	switch {
	case err != nil:
		return models.AccountPassword{}, fmt.Errorf(
			"failed to update account password for account %s, cause: %w", accountID, err,
		)
	case acc.IsNil():
		return models.AccountPassword{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account password for account %s not found", accountID),
		)
	}

	return acc.ToModel(), nil
}
