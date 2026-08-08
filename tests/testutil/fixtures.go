package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/auth-svc/internal/repo/pg"
	"github.com/netbill/auth-svc/pkg/passmanager"
	"github.com/netbill/pgdbx"
	"github.com/stretchr/testify/require"
)

const TestPassword = "Test@pass1"

// CreateUser inserts a full user (user + email + password) directly into the DB.
// Returns the created user and its plain-text password.
func CreateUser(t *testing.T, db *pgdbx.DB, email string) (models.User, models.UserEmail, string) {
	t.Helper()

	ctx := context.Background()

	userRepo := pg.NewUserRepo(db)
	emailRepo := pg.NewEmailRepo(db)
	passwordRepo := pg.NewPasswordRepo(db)

	acc, err := userRepo.Create(ctx, user.RegistrationParams{
		Role: "user",
	})
	require.NoError(t, err)

	em, err := emailRepo.Create(ctx, models.UserEmail{
		UserID: acc.ID,
		Email:  email,
	})
	require.NoError(t, err)

	passMgr := passmanager.New(4)
	hash, err := passMgr.GenerateHash(TestPassword)
	require.NoError(t, err)

	_, err = passwordRepo.Create(ctx, models.UserPassword{
		UserID: acc.ID,
		Hash:   hash,
	})
	require.NoError(t, err)

	return acc, em, TestPassword
}

// CreateSession inserts a session directly into the DB for a given user.
func CreateSession(t *testing.T, db *pgdbx.DB, userID uuid.UUID) models.Session {
	t.Helper()

	sessionRepo := pg.NewSessionRepo(db)
	sess, err := sessionRepo.Create(context.Background(), uuid.New(), userID, "testhash")
	require.NoError(t, err)

	return sess
}
