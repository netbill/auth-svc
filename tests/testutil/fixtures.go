package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
	"github.com/netbill/auth-svc/internal/repo/pg"
	"github.com/netbill/auth-svc/pkg/passmanager"
	"github.com/netbill/pgdbx"
	"github.com/stretchr/testify/require"
)

const TestPassword = "Test@pass1"

// CreateAccount inserts a full account (account + email + password) directly into the DB.
// Returns the created account and its plain-text password.
func CreateAccount(t *testing.T, db *pgdbx.DB, email string) (models.Account, models.AccountEmail, string) {
	t.Helper()

	ctx := context.Background()

	accountRepo := pg.NewAccountRepo(db)
	emailRepo := pg.NewEmailRepo(db)
	passwordRepo := pg.NewPasswordRepo(db)

	acc, err := accountRepo.Create(ctx, account.RegistrationParams{
		Role: "user",
	})
	require.NoError(t, err)

	em, err := emailRepo.Create(ctx, models.AccountEmail{
		AccountID: acc.ID,
		Email:     email,
	})
	require.NoError(t, err)

	passMgr := passmanager.New(4)
	hash, err := passMgr.GenerateHash(TestPassword)
	require.NoError(t, err)

	_, err = passwordRepo.Create(ctx, models.AccountPassword{
		AccountID: acc.ID,
		Hash:      hash,
	})
	require.NoError(t, err)

	return acc, em, TestPassword
}

// CreateSession inserts a session directly into the DB for a given account.
func CreateSession(t *testing.T, db *pgdbx.DB, accountID uuid.UUID) models.Session {
	t.Helper()

	sessionRepo := pg.NewSessionRepo(db)
	sess, err := sessionRepo.Create(context.Background(), uuid.New(), accountID, "testhash")
	require.NoError(t, err)

	return sess
}
