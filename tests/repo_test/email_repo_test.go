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

func newEmailRepo(t *testing.T) (*pg.AccountRepo, *pg.EmailRepo) {
	t.Helper()
	pool := setupDB(t)
	db := pgdbx.NewDB(pool)
	return pg.NewAccountRepo(db), pg.NewEmailRepo(db)
}

func TestEmailRepo_Create(t *testing.T) {
	accRepo, emailRepo := newEmailRepo(t)
	ctx := context.Background()

	acc, err := accRepo.Create(ctx, account.RegistrationParams{Role: "user"})
	require.NoError(t, err)

	email, err := emailRepo.Create(ctx, models.AccountEmail{
		AccountID: acc.ID,
		Email:     "alice@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, acc.ID, email.AccountID)
	assert.Equal(t, "alice@example.com", email.Email)
	assert.False(t, email.Verified)
}

func TestEmailRepo_Create_DuplicateEmail(t *testing.T) {
	accRepo, emailRepo := newEmailRepo(t)
	ctx := context.Background()

	acc1, err := accRepo.Create(ctx, account.RegistrationParams{Role: "user"})
	require.NoError(t, err)
	acc2, err := accRepo.Create(ctx, account.RegistrationParams{Role: "user"})
	require.NoError(t, err)

	_, err = emailRepo.Create(ctx, models.AccountEmail{AccountID: acc1.ID, Email: "dup@example.com"})
	require.NoError(t, err)

	_, err = emailRepo.Create(ctx, models.AccountEmail{AccountID: acc2.ID, Email: "dup@example.com"})
	assert.ErrorIs(t, err, errx.ErrorEmailAlreadyExist)
}

func TestEmailRepo_GetByID_Active(t *testing.T) {
	accRepo, emailRepo := newEmailRepo(t)
	ctx := context.Background()

	acc, err := accRepo.Create(ctx, account.RegistrationParams{Role: "user"})
	require.NoError(t, err)

	_, err = emailRepo.Create(ctx, models.AccountEmail{AccountID: acc.ID, Email: "bob@example.com"})
	require.NoError(t, err)

	got, err := emailRepo.GetByID(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, "bob@example.com", got.Email)
}

func TestEmailRepo_GetByID_NotFound(t *testing.T) {
	_, emailRepo := newEmailRepo(t)

	_, err := emailRepo.GetByID(context.Background(), testutil.RandomUUID())
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}

func TestEmailRepo_GetByID_WithDeletedFilter(t *testing.T) {
	accRepo, emailRepo := newEmailRepo(t)
	ctx := context.Background()

	acc, err := accRepo.Create(ctx, account.RegistrationParams{Role: "user"})
	require.NoError(t, err)

	_, err = emailRepo.Create(ctx, models.AccountEmail{AccountID: acc.ID, Email: "charlie@example.com"})
	require.NoError(t, err)

	// Soft-delete account (cascade triggers soft-delete on email)
	_, err = accRepo.Delete(ctx, acc.ID)
	require.NoError(t, err)

	// Default filter (active only) → not found
	_, err = emailRepo.GetByID(ctx, acc.ID)
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)

	// WithDeleted filter → found
	got, err := emailRepo.GetByID(ctx, acc.ID, account.WithDeleted(account.DeletedFilterAll))
	require.NoError(t, err)
	assert.NotNil(t, got.DeletedAt)
}

func TestEmailRepo_GetByEmail(t *testing.T) {
	accRepo, emailRepo := newEmailRepo(t)
	ctx := context.Background()

	acc, err := accRepo.Create(ctx, account.RegistrationParams{Role: "user"})
	require.NoError(t, err)

	_, err = emailRepo.Create(ctx, models.AccountEmail{AccountID: acc.ID, Email: "dave@example.com"})
	require.NoError(t, err)

	got, err := emailRepo.GetByEmail(ctx, "dave@example.com")
	require.NoError(t, err)
	assert.Equal(t, acc.ID, got.AccountID)
}

func TestEmailRepo_GetByEmail_NotFound(t *testing.T) {
	_, emailRepo := newEmailRepo(t)

	_, err := emailRepo.GetByEmail(context.Background(), "nobody@example.com")
	assert.ErrorIs(t, err, errx.ErrorAccountNotFound)
}
