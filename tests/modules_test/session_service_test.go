package modules_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func registerAndLogin(
	t *testing.T,
	userSvc *user.Service,
	sessionSvc interface {
		LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error)
	},
) (models.User, models.TokensPair, string) {
	t.Helper()
	ctx := context.Background()

	email := testutil.UniqueEmail()
	acc, err := userSvc.Registration(ctx, user.RegistrationParams{
		Email:    email,
		Password: testutil.TestPassword,
		Username: testutil.UniqueUsername(),
		Role:     "user",
	})
	require.NoError(t, err)

	tokens, err := sessionSvc.LoginByEmail(ctx, email, testutil.TestPassword)
	require.NoError(t, err)

	return acc, tokens, email
}

func TestSessionService_LoginByEmail(t *testing.T) {
	db, rc := setup(t)
	userSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	email := testutil.UniqueEmail()
	_, err := userSvc.Registration(ctx, user.RegistrationParams{
		Email:    email,
		Password: testutil.TestPassword,
		Username: testutil.UniqueUsername(),
		Role:     "user",
	})
	require.NoError(t, err)

	tokens, err := sessionSvc.LoginByEmail(ctx, email, testutil.TestPassword)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.Access)
	assert.NotEmpty(t, tokens.Refresh)
	assert.NotEqual(t, uuid.Nil, tokens.SessionID)
}

func TestSessionService_LoginByEmail_WrongPassword(t *testing.T) {
	db, rc := setup(t)
	userSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	email := testutil.UniqueEmail()
	_, err := userSvc.Registration(ctx, user.RegistrationParams{
		Email:    email,
		Password: testutil.TestPassword,
		Username: testutil.UniqueUsername(),
		Role:     "user",
	})
	require.NoError(t, err)

	_, err = sessionSvc.LoginByEmail(ctx, email, "Wrong@pass1")
	assert.Error(t, err)
}

func TestSessionService_LoginByEmail_NotFound(t *testing.T) {
	db, rc := setup(t)
	_, sessionSvc := newServices(t, db, rc)

	_, err := sessionSvc.LoginByEmail(context.Background(), "nobody@example.com", testutil.TestPassword)
	assert.ErrorIs(t, err, errx.ErrorUserNotFound)
}

func TestSessionService_GetMySession(t *testing.T) {
	db, rc := setup(t)
	userSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	acc, tokens, _ := registerAndLogin(t, userSvc, sessionSvc)

	actor := models.UserActor{
		ID:        acc.ID,
		SessionID: tokens.SessionID,
		Role:      acc.Role,
	}

	sess, err := sessionSvc.GetMySession(ctx, actor, tokens.SessionID)
	require.NoError(t, err)
	assert.Equal(t, tokens.SessionID, sess.ID)
	assert.Equal(t, acc.ID, sess.UserID)
}

func TestSessionService_GetMySession_WrongUser(t *testing.T) {
	db, rc := setup(t)
	userSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	_, tokens1, _ := registerAndLogin(t, userSvc, sessionSvc)
	acc2, tokens2, _ := registerAndLogin(t, userSvc, sessionSvc)

	// acc2 tries to read acc1's session
	actor2 := models.UserActor{
		ID:        acc2.ID,
		SessionID: tokens2.SessionID,
		Role:      acc2.Role,
	}

	_, err := sessionSvc.GetMySession(ctx, actor2, tokens1.SessionID)
	assert.ErrorIs(t, err, errx.ErrorSessionNotFound)
}

func TestSessionService_Logout(t *testing.T) {
	db, rc := setup(t)
	userSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	acc, tokens, _ := registerAndLogin(t, userSvc, sessionSvc)

	actor := models.UserActor{
		ID:        acc.ID,
		SessionID: tokens.SessionID,
		Role:      acc.Role,
	}

	err := sessionSvc.Logout(ctx, actor)
	require.NoError(t, err)
	time.Sleep(time.Second)

	// After logout the session should be soft-deleted — GetMySession should ok because we can get deleted session
	session, err := sessionSvc.GetMySession(ctx, actor, tokens.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, session.DeletedAt)
}

func TestSessionService_Refresh(t *testing.T) {
	db, rc := setup(t)
	userSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	_, tokens, _ := registerAndLogin(t, userSvc, sessionSvc)

	// Wait >1s so the new refresh token gets a different ExpiresAt (second-precision JWT)
	time.Sleep(1100 * time.Millisecond)

	newTokens, err := sessionSvc.Refresh(ctx, tokens.Refresh)
	require.NoError(t, err)
	assert.NotEmpty(t, newTokens.Access)
	assert.NotEmpty(t, newTokens.Refresh)
	assert.Equal(t, tokens.SessionID, newTokens.SessionID)

	// Old refresh token should no longer work
	_, err = sessionSvc.Refresh(ctx, tokens.Refresh)
	assert.ErrorIs(t, err, errx.ErrorSessionTokenMismatch)
}

func TestSessionService_DeleteMySession(t *testing.T) {
	db, rc := setup(t)
	userSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	acc, tokens, email := registerAndLogin(t, userSvc, sessionSvc)

	// Login again as acc to get a second session
	tokens2, err := sessionSvc.LoginByEmail(ctx, email, testutil.TestPassword)
	require.NoError(t, err)

	actor := models.UserActor{
		ID:        acc.ID,
		SessionID: tokens.SessionID,
		Role:      acc.Role,
	}

	// Delete the second session
	err = sessionSvc.DeleteMySession(ctx, actor, tokens2.SessionID)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// First session should still work
	_, err = sessionSvc.GetMySession(ctx, actor, tokens.SessionID)
	require.NoError(t, err)

	// Second session should be gone
	session, err := sessionSvc.GetMySession(ctx, actor, tokens2.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, session.DeletedAt)
}

func TestSessionService_DeleteMySessions(t *testing.T) {
	db, rc := setup(t)
	userSvc, sessionSvc := newServices(t, db, rc)
	ctx := context.Background()

	acc, tokens, email := registerAndLogin(t, userSvc, sessionSvc)

	// Create additional sessions
	tokens2, err := sessionSvc.LoginByEmail(ctx, email, testutil.TestPassword)
	require.NoError(t, err)
	tokens3, err := sessionSvc.LoginByEmail(ctx, email, testutil.TestPassword)
	require.NoError(t, err)

	// Use the first session as the actor for the DeleteMySessions call
	actor := models.UserActor{
		ID:        acc.ID,
		SessionID: tokens.SessionID,
		Role:      acc.Role,
	}

	err = sessionSvc.DeleteMySessions(ctx, actor)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// All sessions should be deleted
	session, err := sessionSvc.GetMySession(ctx, actor, tokens.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, session.DeletedAt)
	session, err = sessionSvc.GetMySession(ctx, actor, tokens2.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, session.DeletedAt)
	session, err = sessionSvc.GetMySession(ctx, actor, tokens3.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, session.DeletedAt)
}
