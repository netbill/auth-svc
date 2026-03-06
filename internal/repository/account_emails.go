package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
)

type AccountEmailRow struct {
	AccountID uuid.UUID `db:"account_id"`
	Email     string    `db:"email"`
	Verified  bool      `db:"verified"`
	Version   int32     `db:"version"`
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
		Version:   a.Version,
		UpdatedAt: a.UpdatedAt,
		CreatedAt: a.CreatedAt,
	}
}

type AccountEmailsQ interface {
	New() AccountEmailsQ
	Insert(ctx context.Context, input AccountEmailRow) (AccountEmailRow, error)

	Get(ctx context.Context) (AccountEmailRow, error)
	Select(ctx context.Context) ([]AccountEmailRow, error)

	UpdateOne(ctx context.Context) (AccountEmailRow, error)

	UpdateEmail(email string) AccountEmailsQ
	UpdateVerified(verified bool) AccountEmailsQ

	Delete(ctx context.Context) error

	FilterAccountID(accountID uuid.UUID) AccountEmailsQ
	FilterEmail(email string) AccountEmailsQ

	Exists(ctx context.Context) (bool, error)
}

type AccountEmailRepo struct {
	query AccountEmailsQ
}

func NewAccountEmailRepo(query AccountEmailsQ) *AccountEmailRepo {
	return &AccountEmailRepo{
		query: query,
	}
}

func (r *AccountEmailRepo) Create(ctx context.Context, params models.AccountEmail) error {
	_, err := r.query.New().Insert(ctx, AccountEmailRow{
		AccountID: params.AccountID,
		Email:     params.Email,
	})
	if err != nil {
		return fmt.Errorf(
			"failed to create account email for account %s, cause: %w", params.AccountID, err,
		)
	}

	return nil
}

func (r *AccountEmailRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exist, err := r.query.New().FilterEmail(email).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence by email %s, cause: %w", email, err)
	}

	return exist, nil
}

func (r *AccountEmailRepo) GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error) {
	row, err := r.query.New().FilterAccountID(accountID).Get(ctx)
	switch {
	case err != nil:
		return models.AccountEmail{}, fmt.Errorf(
			"failed to get account email for account %s, cause: %w", accountID, err,
		)
	case row.IsNil():
		return models.AccountEmail{}, errx.ErrorAccountNotFound.Raise(err)
	}

	return row.ToModel(), nil
}
