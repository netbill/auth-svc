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

func newAccountRepo(t *testing.T) *pg.AccountRepo {
	t.Helper()
	pool := setupDB(t)
	return pg.NewAccountRepo(pgdbx.NewDB(pool))
}

func TestAccountRepo_Create(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	acc, err := repo.Create(ctx, account.RegistrationParams{
		Username: "alice",
		Role:     "user",
	})
	require.NoError(t, err)
	assert.Equal(t, "alice", acc.Username)
	assert.Equal(t, "user", acc.Role)
	assert.NotEmpty(t, acc.ID)
	assert.Nil(t, acc.DeletedAt)
}

func TestAccountRepo_Create_DuplicateUsername(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, account.RegistrationParams{Username: "alice", Role: "user"})
	require.NoError(t, err)

	_, err = repo.Create(ctx, account.RegistrationParams{Username: "alice", Role: "user"})
	assert.ErrorIs(t, err, errx.ErrorUsernameAlreadyTaken)
}

func TestAccountRepo_GetByID(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, account.RegistrationParams{Username: "bob", Role: "user"})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "bob", got.Username)
}

func TestAccountRepo_GetByID_NotFound(t *testing.T) {
	repo := newAccountRepo(t)

	_, err := repo.GetByID(context.Background(), testutil.RandomUUID())
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestAccountRepo_GetByUsername(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, account.RegistrationParams{Username: "charlie", Role: "user"})
	require.NoError(t, err)

	got, err := repo.GetByUsername(ctx, "charlie")
	require.NoError(t, err)
	assert.Equal(t, "charlie", got.Username)
}

func TestAccountRepo_GetByUsername_NotFound(t *testing.T) {
	repo := newAccountRepo(t)

	_, err := repo.GetByUsername(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestAccountRepo_UpdateUsername(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, account.RegistrationParams{Username: "dave", Role: "user"})
	require.NoError(t, err)

	updated, err := repo.UpdateUsername(ctx, created.ID, "dave_updated", created.Version)
	require.NoError(t, err)
	assert.Equal(t, "dave_updated", updated.Username)
	assert.Equal(t, created.Version+1, updated.Version)
}

func TestAccountRepo_UpdateUsername_Conflict(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	acc1, err := repo.Create(ctx, account.RegistrationParams{Username: "eve", Role: "user"})
	require.NoError(t, err)

	_, err = repo.Create(ctx, account.RegistrationParams{Username: "frank", Role: "user"})
	require.NoError(t, err)

	// Try to rename eve → frank (taken)
	_, err = repo.UpdateUsername(ctx, acc1.ID, "frank", acc1.Version)
	assert.ErrorIs(t, err, errx.ErrorUsernameAlreadyTaken)
}

func TestAccountRepo_UpdateUsername_StaleVersion(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, account.RegistrationParams{Username: "grace", Role: "user"})
	require.NoError(t, err)

	// Update with wrong version — should return not found (0 rows affected)
	_, err = repo.UpdateUsername(ctx, created.ID, "grace_new", created.Version+99)
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestAccountRepo_Delete(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, account.RegistrationParams{Username: "henry", Role: "user"})
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

	created, err := repo.Create(ctx, account.RegistrationParams{Username: "ivan", Role: "user"})
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

// Ensure deleted account username is anonymized.
func TestAccountRepo_Delete_AnonymizesUsername(t *testing.T) {
	repo := newAccountRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, account.RegistrationParams{Username: "judy", Role: "user"})
	require.NoError(t, err)

	deleted, err := repo.Delete(ctx, created.ID)
	require.NoError(t, err)
	assert.NotEqual(t, "judy", deleted.Username, "username should be anonymized on delete")

	// Should be possible to create a new account with the same username after deletion.
	_, err = repo.Create(ctx, account.RegistrationParams{Username: "judy", Role: "user"})
	assert.NoError(t, err)
}

// Helper — generate a random UUID that does not exist in the DB.
func init() {
	_ = models.Account{} // ensure models imported
}
