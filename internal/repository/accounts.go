package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/core/modules/auth"
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

func (r *Repository) CreateAccount(ctx context.Context, params auth.RegistrationParams) (models.Account, error) {
	accountID := uuid.New()

	acc, err := r.AccountsQ.New().Insert(ctx, AccountRow{
		ID:       accountID,
		Username: params.Username,
		Role:     params.Role,
	})
	if err != nil {
		return models.Account{}, fmt.Errorf("failed to insert account, cause: %w", err)
	}

	if _, err = r.AccountEmailsQ.New().Insert(ctx, AccountEmailRow{
		AccountID: accountID,
		Email:     params.Email,
		Verified:  false,
	}); err != nil {
		return models.Account{}, fmt.Errorf("failed to insert account email, cause: %w", err)
	}

	if _, err = r.AccountPassQ.New().Insert(ctx, AccountPasswordRow{
		AccountID: accountID,
		Hash:      params.GetPassHash(),
	}); err != nil {
		return models.Account{}, fmt.Errorf("failed to insert account password, cause: %w", err)
	}

	return acc.ToModel(), nil
}

func (r *Repository) GetAccountByID(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	row, err := r.AccountsQ.New().FilterID(accountID).Get(ctx)
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

func (r *Repository) ExistsAccountByID(ctx context.Context, accountID uuid.UUID) (bool, error) {
	exist, err := r.AccountsQ.New().FilterID(accountID).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence by id %s, cause: %w", accountID, err)
	}

	return exist, nil
}

func (r *Repository) GetAccountByUsername(ctx context.Context, username string) (models.Account, error) {
	row, err := r.AccountsQ.New().FilterUsername(username).Get(ctx)
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

func (r *Repository) GetAccountByEmail(ctx context.Context, email string) (models.Account, error) {
	row, err := r.AccountsQ.New().FilterEmail(email).Get(ctx)
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

func (r *Repository) ExistsAccountByUsername(ctx context.Context, username string) (bool, error) {
	exist, err := r.AccountsQ.New().FilterUsername(username).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence by username %s, cause: %w", username, err)
	}

	return exist, nil
}

func (r *Repository) UpdateAccountUsername(
	ctx context.Context,
	accountID uuid.UUID,
	username string,
) (models.Account, error) {
	row, err := r.AccountsQ.New().
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

func (r *Repository) DeleteAccount(ctx context.Context, accountID uuid.UUID) error {
	err := r.AccountsQ.New().FilterID(accountID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete account %s, cause: %w", accountID, err)
	}

	return nil
}
