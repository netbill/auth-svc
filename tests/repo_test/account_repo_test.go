package repo_test

import (
	"context"
	"testing"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/modules/account"
	"github.com/netbill/auth-svc/internal/repo/pg"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/netbill/pgdbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAccountRepo(t *testing.T) *pg.AccountRepo {
	t.Helper()
	pool := setupDB(t)
	return pg.NewAccountRepo(pgdbx.NewDB(pool))
}

func TestAccountRepo_Create(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	acc, err := repo.Create(ctx, account.RegistrationParams{
		Role: "user",
	})
	require.NoError(t, err)
	assert.Equal(t, "user", acc.Role)
	assert.NotEmpty(t, acc.ID)
	assert.Nil(t, acc.DeletedAt)
}

func TestAccountRepo_GetByID(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, account.RegistrationParams{Role: "user"})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestAccountRepo_GetByID_NotFound(t *testing.T) {
	repo := newAccountRepo(t)

	_, err := repo.GetByID(context.Background(), testutil.RandomUUID())
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestAccountRepo_Delete(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, account.RegistrationParams{Role: "user"})
	require.NoError(t, err)

	deleted, err := repo.Delete(ctx, created.ID)
	require.NoError(t, err)
	assert.NotNil(t, deleted.DeletedAt)

	// Account is soft-deleted — GetByID returns not found
	_, err = repo.GetByID(ctx, created.ID)
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestAccountRepo_Delete_AlreadyDeleted(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, account.RegistrationParams{Role: "user"})
	require.NoError(t, err)

	_, err = repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = repo.Delete(ctx, created.ID)
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestAccountRepo_Delete_NotFound(t *testing.T) {
	repo := newAccountRepo(t)

	_, err := repo.Delete(context.Background(), testutil.RandomUUID())
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}
