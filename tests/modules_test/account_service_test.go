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
		Username: testutil.UniqueUsername(),
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, acc.ID)
	assert.Equal(t, "user", acc.Role)
}

func TestAccountService_Registration_DuplicateUsername(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	username := testutil.UniqueUsername()

	_, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Username: username,
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)

	_, err = accountSvc.Registration(ctx, account.RegistrationParams{
		Username: username,
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	assert.ErrorIs(t, err, errx.ErrorUsernameAlreadyTaken)
}

func TestAccountService_Registration_DuplicateEmail(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	email := testutil.UniqueEmail()

	_, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Username: testutil.UniqueUsername(),
		Email:    email,
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)

	_, err = accountSvc.Registration(ctx, account.RegistrationParams{
		Username: testutil.UniqueUsername(),
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
		Username: testutil.UniqueUsername(),
		Email:    testutil.UniqueEmail(),
		Password: "weak",
		Role:     "user",
	})
	assert.ErrorIs(t, err, errx.ErrorPasswordIsNotAllowed)
}

func TestAccountService_Registration_InvalidUsername(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	_, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Username: "a b", // space not allowed
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	assert.ErrorIs(t, err, errx.ErrorUsernameIsNotAllowed)
}

func TestAccountService_GetMyAccountByID(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	username := testutil.UniqueUsername()
	created, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Username: username,
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)

	actor := models.AccountActor{ID: created.ID, Role: created.Role}

	got, err := accountSvc.GetMyAccountByID(ctx, actor)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, username, got.Username)
}

func TestAccountService_UpdateUsername(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	created, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Username: testutil.UniqueUsername(),
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)

	actor := models.AccountActor{ID: created.ID, Role: created.Role}
	newName := testutil.UniqueUsername()

	updated, err := accountSvc.UpdateUsername(ctx, actor, newName)
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Username)
}

func TestAccountService_UpdateUsername_SameName(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	username := testutil.UniqueUsername()
	created, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Username: username,
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)

	actor := models.AccountActor{ID: created.ID, Role: created.Role}

	// Update to the same username — should succeed and return the same account
	got, err := accountSvc.UpdateUsername(ctx, actor, username)
	require.NoError(t, err)
	assert.Equal(t, username, got.Username)
	assert.Equal(t, created.Version, got.Version, "version should not increment on no-op update")
}

func TestAccountService_UpdateUsername_InvalidChars(t *testing.T) {
	db, rc := setup(t)
	accountSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	created, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Username: testutil.UniqueUsername(),
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Role:     "user",
	})
	require.NoError(t, err)

	actor := models.AccountActor{ID: created.ID, Role: created.Role}

	_, err = accountSvc.UpdateUsername(ctx, actor, "bad name!")
	assert.ErrorIs(t, err, errx.ErrorUsernameIsNotAllowed)
}

func TestAccountService_UpdatePassword(t *testing.T) {
	db, rc := setup(t)
	accountSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	email := testutil.UniqueEmail()
	created, err := accountSvc.Registration(ctx, account.RegistrationParams{
		Username: testutil.UniqueUsername(),
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
