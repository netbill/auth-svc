package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/domain/models"
	"github.com/netbill/auth-svc/internal/domain/modules/auth"
	"github.com/netbill/auth-svc/internal/repository/pgdb"
)

func (r Repository) CreateAccount(ctx context.Context, params auth.CreateAccountParams) (models.Account, error) {
	now := time.Now().UTC()
	accountID := uuid.New()

	acc := pgdb.Account{
		ID:                accountID,
		Username:          params.Username,
		Role:              params.Role,
		Status:            models.AccountStatusActive,
		CreatedAt:         now,
		UpdatedAt:         now,
		UsernameUpdatedAt: now,
	}

	account := acc.ToModel()

	err := r.accountsQ().Insert(ctx, acc)
	if err != nil {
		return models.Account{}, err
	}

	emailRow := pgdb.AccountEmail{
		AccountID: accountID,
		Email:     params.Email,
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = r.emailsQ().Insert(ctx, emailRow)
	if err != nil {
		return models.Account{}, err
	}

	passwordRow := pgdb.AccountPassword{
		AccountID: accountID,
		Hash:      params.PasswordHash,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = r.passwordsQ().Insert(ctx, passwordRow)
	if err != nil {
		return models.Account{}, err
	}

	return account, err
}

func (r Repository) GetAccountByID(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	acc, err := r.accountsQ().FilterID(accountID).Get(ctx)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.Account{}, nil
	case err != nil:
		return models.Account{}, err
	}

	return acc.ToModel(), nil
}

func (r Repository) GetAccountByUsername(ctx context.Context, username string) (models.Account, error) {
	acc, err := r.accountsQ().FilterUsername(username).Get(ctx)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.Account{}, nil
	case err != nil:
		return models.Account{}, err
	}

	return acc.ToModel(), nil
}

func (r Repository) GetAccountByEmail(ctx context.Context, email string) (models.Account, error) {
	acc, err := r.accountsQ().FilterEmail(email).Get(ctx)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.Account{}, nil
	case err != nil:
		return models.Account{}, err
	}

	return acc.ToModel(), nil
}

func (r Repository) UpdateAccountUsername(ctx context.Context, accountID uuid.UUID, newUsername string) (models.Account, error) {
	var account models.Account

	accs, err := r.accountsQ().
		FilterID(accountID).
		UpdateUsername(newUsername, time.Now().UTC()).
		Update(ctx)
	if err != nil {
		return models.Account{}, err
	}

	if len(accs) == 1 {
		account = accs[0].ToModel()
	} else {
		return models.Account{}, fmt.Errorf("expected to update 1 account, updated %d", len(accs))
	}

	return account, nil
}

func (r Repository) UpdateAccountStatus(ctx context.Context, accountID uuid.UUID, status string) (models.Account, error) {
	accs, err := r.accountsQ().
		FilterID(accountID).
		UpdateStatus(status).
		Update(ctx)
	if err != nil {
		return models.Account{}, err
	}

	if len(accs) != 1 {
		return models.Account{}, fmt.Errorf("expected to update 1 account, updated %d", len(accs))
	}
	return accs[0].ToModel(), nil
}

func (r Repository) GetAccountEmail(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error) {
	acc, err := r.emailsQ().FilterAccountID(accountID).Get(ctx)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.AccountEmail{}, nil
	case err != nil:
		return models.AccountEmail{}, err
	}

	return acc.ToModel(), nil
}

func (r Repository) UpdateAccountEmailVerification(
	ctx context.Context,
	accountID uuid.UUID,
	verified bool,
) (models.AccountEmail, error) {
	accs, err := r.emailsQ().
		FilterAccountID(accountID).
		UpdateVerified(verified).
		Update(ctx)
	if err != nil {
		return models.AccountEmail{}, err
	}

	if len(accs) != 1 {
		return models.AccountEmail{}, fmt.Errorf("expected to update 1 account, updated %d", len(accs))
	}
	return accs[0].ToModel(), nil
}

func (r Repository) GetAccountPassword(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error) {
	acc, err := r.passwordsQ().FilterAccountID(accountID).Get(ctx)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.AccountPassword{}, nil
	case err != nil:
		return models.AccountPassword{}, err
	}

	return acc.ToModel(), nil
}

func (r Repository) UpdateAccountPassword(
	ctx context.Context,
	accountID uuid.UUID,
	passwordHash string,
) (models.AccountPassword, error) {
	accs, err := r.passwordsQ().
		FilterAccountID(accountID).
		UpdateHash(passwordHash).
		Update(ctx)
	if err != nil {
		return models.AccountPassword{}, err
	}

	if len(accs) != 1 {
		return models.AccountPassword{}, fmt.Errorf("expected to update 1 account, updated %d", len(accs))
	}

	return accs[0].ToModel(), nil
}

func (r Repository) DeleteAccount(ctx context.Context, accountID uuid.UUID) error {
	return r.accountsQ().FilterID(accountID).Delete(ctx)
}
