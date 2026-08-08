package modules_test

import (
	"context"
	"testing"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountService_Registration(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	acc, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, acc.ID)
	assert.Equal(t, "user", acc.Role)
}

func TestAccountService_Registration_DuplicateEmail(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	email := testutil.UniqueEmail()

	_, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Email:    email,
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)

	_, err = accountSvc.Registration(ctx, account.RegistrationParams{
		Email:    email,
		Password: testutil.TestPassword,
		Role:     "user",
	})
	assert.ErrorIs(t, err, errx.ErrorEmailAlreadyExist)
}

func TestAccountService_Registration_WeakPassword(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	_, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Email:    testutil.UniqueEmail(),
		Password: "weak",
		Role:     "user",
	})
	assert.ErrorIs(t, err, errx.ErrorPasswordIsNotAllowed)
}

func TestAccountService_GetMyAccountByID(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	created, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)

	actor := models.AccountActor{ID: created.ID, Role: created.Role}

	got, err := accountSvc.GetMyAccountByID(ctx, actor)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestAccountService_UpdatePassword(t *testing.T) {
	db, rc := setup(t)
	accountSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	email := testutil.UniqueEmail()
	created, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Email:    email,
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)

	// Login to get a valid session
	tokens, err := sessionSvc.LoginByEmail(ctx, email, testutil.TestPassword)
	require.NoError(t, err)

	actor := models.AccountActor{
		ID:        created.ID,
		SessionID: tokens.SessionID,
		Role:      created.Role,
	}

	const newPass = "NewPass@1234"
	err = accountSvc.UpdatePassword(ctx, actor, testutil.TestPassword, newPass)
	require.NoError(t, err)

	// Verify new password works for login
	_, err = sessionSvc.LoginByEmail(ctx, email, newPass)
	require.NoError(t, err)
}
