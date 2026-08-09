package modules_test

import (
	"context"
	"testing"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_Registration(t *testing.T) {
	db, rc := setup(t)
	userSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	acc, err := userSvc.Registration(ctx, user.RegistrationParams{
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Username: testutil.UniqueUsername(),
		Role:     "user",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, acc.ID)
	assert.Equal(t, "user", acc.Role)
}

func TestUserService_Registration_DuplicateEmail(t *testing.T) {
	db, rc := setup(t)
	userSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	email := testutil.UniqueEmail()

	_, err := userSvc.Registration(ctx, user.RegistrationParams{
		Email:    email,
		Password: testutil.TestPassword,
		Username: testutil.UniqueUsername(),
		Role:     "user",
	})
	require.NoError(t, err)

	_, err = userSvc.Registration(ctx, user.RegistrationParams{
		Email:    email,
		Password: testutil.TestPassword,
		Username: testutil.UniqueUsername(),
		Role:     "user",
	})
	assert.ErrorIs(t, err, errx.ErrorEmailAlreadyExist)
}

func TestUserService_Registration_WeakPassword(t *testing.T) {
	db, rc := setup(t)
	userSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	_, err := userSvc.Registration(ctx, user.RegistrationParams{
		Email:    testutil.UniqueEmail(),
		Password: "weak",
		Username: testutil.UniqueUsername(),
		Role:     "user",
	})
	assert.ErrorIs(t, err, errx.ErrorPasswordIsNotAllowed)
}

func TestUserService_GetMyUserByID(t *testing.T) {
	db, rc := setup(t)
	userSvc, _ := newServices(t, db, rc)
	ctx := context.Background()

	created, err := userSvc.Registration(ctx, user.RegistrationParams{
		Email:    testutil.UniqueEmail(),
		Password: testutil.TestPassword,
		Username: testutil.UniqueUsername(),
		Role:     "user",
	})
	require.NoError(t, err)

	actor := models.UserActor{ID: created.ID, Role: created.Role}

	got, err := userSvc.GetMyUserByID(ctx, actor)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestUserService_UpdatePassword(t *testing.T) {
	db, rc := setup(t)
	userSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	email := testutil.UniqueEmail()
	created, err := userSvc.Registration(ctx, user.RegistrationParams{
		Email:    email,
		Password: testutil.TestPassword,
		Username: testutil.UniqueUsername(),
		Role:     "user",
	})
	require.NoError(t, err)

	// Login to get a valid session
	tokens, err := sessionSvc.LoginByEmail(ctx, email, testutil.TestPassword)
	require.NoError(t, err)

	actor := models.UserActor{
		ID:        created.ID,
		SessionID: tokens.SessionID,
		Role:      created.Role,
	}

	const newPass = "NewPass@1234"
	err = userSvc.UpdatePassword(ctx, actor, testutil.TestPassword, newPass)
	require.NoError(t, err)

	// Verify new password works for login
	_, err = sessionSvc.LoginByEmail(ctx, email, newPass)
	require.NoError(t, err)
}
