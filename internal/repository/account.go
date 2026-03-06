package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/auth"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
)

type AccountRow struct {
	ID        uuid.UUID `db:"id"`
	Username  string    `db:"username"`
	Role      string    `db:"role"`
	Version   int32     `db:"version"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (a AccountRow) IsNil() bool {
	return a.ID == uuid.Nil
}

func (a AccountRow) ToModel() models.Account {
	return models.Account{
		ID:        a.ID,
		Username:  a.Username,
		Role:      a.Role,
		Version:   a.Version,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

type AccountsQ interface {
	New() AccountsQ
	Insert(ctx context.Context, input AccountRow) (AccountRow, error)

	Get(ctx context.Context) (AccountRow, error)
	Select(ctx context.Context) ([]AccountRow, error)

	UpdateOne(ctx context.Context) (AccountRow, error)

	UpdateUsername(username string) AccountsQ
	UpdateRole(role string) AccountsQ

	Delete(ctx context.Context) error

	FilterID(accountID uuid.UUID) AccountsQ
	FilterEmail(email string) AccountsQ
	FilterUsername(username string) AccountsQ
	FilterVersion(version int32) AccountsQ

	Exists(ctx context.Context) (bool, error)
}

type AccountRepo struct {
	query AccountsQ
}

func NewAccountRepo(query AccountsQ) *AccountRepo {
	return &AccountRepo{
		query: query,
	}
}

func (r *AccountRepo) Create(ctx context.Context, params auth.RegistrationParams) (models.Account, error) {
	accountID := uuid.New()

	acc, err := r.query.New().Insert(ctx, AccountRow{
		ID:       accountID,
		Username: params.Username,
		Role:     params.Role,
	})
	if err != nil {
		return models.Account{}, fmt.Errorf("failed to insert account, cause: %w", err)
	}

	return acc.ToModel(), nil
}

func (r *AccountRepo) GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	row, err := r.query.New().FilterID(accountID).Get(ctx)
	switch {
	case err != nil:
		return models.Account{}, fmt.Errorf("failed to get account, cause: %w", err)
	case row.IsNil():
		return models.Account{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account with id %s not found", accountID),
		)
	}

	return row.ToModel(), nil
}

func (r *AccountRepo) ExistsByID(ctx context.Context, accountID uuid.UUID) (bool, error) {
	exist, err := r.query.New().FilterID(accountID).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence by id %s, cause: %w", accountID, err)
	}

	return exist, nil
}

func (r *AccountRepo) GetByUsername(ctx context.Context, username string) (models.Account, error) {
	row, err := r.query.New().FilterUsername(username).Get(ctx)
	switch {
	case err != nil:
		return models.Account{}, fmt.Errorf("failed to get account by username, cause: %w", err)
	case row.IsNil():
		return models.Account{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account with username %s not found", username),
		)
	}

	return row.ToModel(), nil
}

func (r *AccountRepo) GetByEmail(ctx context.Context, email string) (models.Account, error) {
	row, err := r.query.New().FilterEmail(email).Get(ctx)
	switch {
	case err != nil:
		return models.Account{}, fmt.Errorf("failed to get account by email, cause: %w", err)
	case row.IsNil():
		return models.Account{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account with email %s not found", email),
		)
	}

	return row.ToModel(), nil
}

func (r *AccountRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	exist, err := r.query.New().FilterUsername(username).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence by username %s, cause: %w", username, err)
	}

	return exist, nil
}

func (r *AccountRepo) UpdateUsername(
	ctx context.Context,
	accountID uuid.UUID,
	username string,
) (models.Account, error) {
	row, err := r.query.New().
		FilterID(accountID).
		UpdateUsername(username).
		UpdateOne(ctx)
	if err != nil {
		return models.Account{}, fmt.Errorf(
			"failed to update account username for account %s, cause: %w", accountID, err,
		)
	}

	return row.ToModel(), nil
}

func (r *AccountRepo) Delete(ctx context.Context, accountID uuid.UUID) error {
	err := r.query.New().FilterID(accountID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete account %s, cause: %w", accountID, err)
	}

	return nil
}
