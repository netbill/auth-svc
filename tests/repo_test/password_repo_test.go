package repo_test

import (
	"context"
	"testing"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
	"github.com/netbill/auth-svc/internal/repo/pg"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/netbill/pgdbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPasswordRepo(t *testing.T) (*pg.AccountRepo, *pg.PasswordRepo) {
	t.Helper()
	pool := setupDB(t)
	db := pgdbx.NewDB(pool)
	return pg.NewAccountRepo(db), pg.NewPasswordRepo(db)
}

func createAccountForPassword(t *testing.T, accRepo *pg.AccountRepo) models.Account {
	t.Helper()
	acc, err := accRepo.Create(context.Background(), account.RegistrationParams{
		Role: "user",
	})
	require.NoError(t, err)
	return acc
}

func TestPasswordRepo_Create(t *testing.T) {
	accRepo, passRepo := newPasswordRepo(t)
	ctx := context.Background()

	acc := createAccountForPassword(t, accRepo)

	pass, err := passRepo.Create(ctx, models.AccountPassword{
		AccountID: acc.ID,
		Hash:      "somehash",
	})
	require.NoError(t, err)
	assert.Equal(t, acc.ID, pass.AccountID)
	assert.Equal(t, "somehash", pass.Hash)
	assert.Nil(t, pass.DeletedAt)
}

func TestPasswordRepo_GetByID(t *testing.T) {
	accRepo, passRepo := newPasswordRepo(t)
	ctx := context.Background()

	acc := createAccountForPassword(t, accRepo)

	_, err := passRepo.Create(ctx, models.AccountPassword{
		AccountID: acc.ID,
		Hash:      "testhash",
	})
	require.NoError(t, err)

	got, err := passRepo.GetByID(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, acc.ID, got.AccountID)
	assert.Equal(t, "testhash", got.Hash)
}

func TestPasswordRepo_GetByID_NotFound(t *testing.T) {
	_, passRepo := newPasswordRepo(t)

	_, err := passRepo.GetByID(context.Background(), testutil.RandomUUID())
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestPasswordRepo_UpdatePassword(t *testing.T) {
	accRepo, passRepo := newPasswordRepo(t)
	ctx := context.Background()

	acc := createAccountForPassword(t, accRepo)

	_, err := passRepo.Create(ctx, models.AccountPassword{
		AccountID: acc.ID,
		Hash:      "oldhash",
	})
	require.NoError(t, err)

	updated, err := passRepo.UpdatePassword(ctx, acc.ID, "newhash")
	require.NoError(t, err)
	assert.Equal(t, acc.ID, updated.AccountID)
	assert.Equal(t, "newhash", updated.Hash)

	// Verify update persisted
	got, err := passRepo.GetByID(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, "newhash", got.Hash)
}

func TestPasswordRepo_UpdatePassword_VersionIncremented(t *testing.T) {
	accRepo, passRepo := newPasswordRepo(t)
	ctx := context.Background()

	acc := createAccountForPassword(t, accRepo)

	original, err := passRepo.Create(ctx, models.AccountPassword{
		AccountID: acc.ID,
		Hash:      "v1hash",
	})
	require.NoError(t, err)

	updated, err := passRepo.UpdatePassword(ctx, acc.ID, "v2hash")
	require.NoError(t, err)
	assert.Equal(t, original.Version+1, updated.Version)
}

func TestPasswordRepo_UpdatePassword_NotFound(t *testing.T) {
	_, passRepo := newPasswordRepo(t)

	_, err := passRepo.UpdatePassword(context.Background(), testutil.RandomUUID(), "hash")
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestPasswordRepo_Delete(t *testing.T) {
	accRepo, passRepo := newPasswordRepo(t)
	ctx := context.Background()

	acc := createAccountForPassword(t, accRepo)

	_, err := passRepo.Create(ctx, models.AccountPassword{
		AccountID: acc.ID,
		Hash:      "hash",
	})
	require.NoError(t, err)

	err = passRepo.Delete(ctx, acc.ID)
	require.NoError(t, err)

	// After soft-delete, GetByID filters deleted_at IS NULL → not found
	_, err = passRepo.GetByID(ctx, acc.ID)
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestPasswordRepo_Delete_NotFound(t *testing.T) {
	_, passRepo := newPasswordRepo(t)

	err := passRepo.Delete(context.Background(), testutil.RandomUUID())
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestPasswordRepo_Delete_AlreadyDeleted(t *testing.T) {
	accRepo, passRepo := newPasswordRepo(t)
	ctx := context.Background()

	acc := createAccountForPassword(t, accRepo)

	_, err := passRepo.Create(ctx, models.AccountPassword{
		AccountID: acc.ID,
		Hash:      "hash",
	})
	require.NoError(t, err)

	err = passRepo.Delete(ctx, acc.ID)
	require.NoError(t, err)

	// Second delete should return not found (already soft-deleted)
	err = passRepo.Delete(ctx, acc.ID)
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}
