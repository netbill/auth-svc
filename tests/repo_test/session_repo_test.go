package repo_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/auth-svc/internal/repo/pg"
	"github.com/netbill/pgdbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSessionRepos(t *testing.T) (*pg.UserRepo, *pg.SessionRepo) {
	t.Helper()
	pool := setupDB(t)
	db := pgdbx.NewDB(pool)
	return pg.NewUserRepo(db), pg.NewSessionRepo(db)
}

func createUserForSession(t *testing.T, accRepo *pg.UserRepo) uuid.UUID {
	t.Helper()
	acc, err := accRepo.Create(context.Background(), user.RegistrationParams{
		Role: "user",
	})
	require.NoError(t, err)
	return acc.ID
}

func TestSessionRepo_Create(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)
	sessionID := uuid.New()

	sess, err := sessRepo.Create(ctx, sessionID, userID, "hashtoken")
	require.NoError(t, err)
	assert.Equal(t, sessionID, sess.ID)
	assert.Equal(t, userID, sess.UserID)
	assert.Nil(t, sess.DeletedAt)
}

func TestSessionRepo_GetByID(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)
	sessionID := uuid.New()

	_, err := sessRepo.Create(ctx, sessionID, userID, "hashtoken")
	require.NoError(t, err)

	got, err := sessRepo.GetByID(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, sessionID, got.ID)
	assert.Equal(t, userID, got.UserID)
}

func TestSessionRepo_GetByID_NotFound(t *testing.T) {
	_, sessRepo := newSessionRepos(t)

	_, err := sessRepo.GetByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, errx.ErrorSessionNotFound)
}

func TestSessionRepo_GetForUser(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)
	sessionID := uuid.New()

	_, err := sessRepo.Create(ctx, sessionID, userID, "hashtoken")
	require.NoError(t, err)

	got, err := sessRepo.GetForUser(ctx, userID, sessionID)
	require.NoError(t, err)
	assert.Equal(t, sessionID, got.ID)
}

func TestSessionRepo_GetForUser_WrongUser(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	acc1 := createUserForSession(t, accRepo)
	acc2 := createUserForSession(t, accRepo)
	sessionID := uuid.New()

	_, err := sessRepo.Create(ctx, sessionID, acc1, "hashtoken")
	require.NoError(t, err)

	// Try to get acc1's session as acc2
	_, err = sessRepo.GetForUser(ctx, acc2, sessionID)
	assert.ErrorIs(t, err, errx.ErrorSessionNotFound)
}

func TestSessionRepo_GetListForUser(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)

	// Create 3 sessions
	for i := 0; i < 3; i++ {
		_, err := sessRepo.Create(ctx, uuid.New(), userID, uuid.New().String())
		require.NoError(t, err)
	}

	page, err := sessRepo.GetListForUser(ctx, userID,
		session.WithLimit(10),
		session.WithOffset(0),
		session.WithDeleted(session.DeletedFilterActive),
	)
	require.NoError(t, err)
	assert.Equal(t, uint(3), page.Total)
	assert.Len(t, page.Data, 3)
}

func TestSessionRepo_GetListForUser_FilterDeleted(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)

	s1, err := sessRepo.Create(ctx, uuid.New(), userID, uuid.New().String())
	require.NoError(t, err)
	_, err = sessRepo.Create(ctx, uuid.New(), userID, uuid.New().String())
	require.NoError(t, err)

	// Delete one session
	err = sessRepo.Delete(ctx, s1.ID)
	require.NoError(t, err)

	// Filter active only
	pageActive, err := sessRepo.GetListForUser(ctx, userID,
		session.WithLimit(10),
		session.WithDeleted(session.DeletedFilterActive),
	)
	require.NoError(t, err)
	assert.Equal(t, uint(1), pageActive.Total)

	// Filter deleted only
	pageDeleted, err := sessRepo.GetListForUser(ctx, userID,
		session.WithLimit(10),
		session.WithDeleted(session.DeletedFilterDeleted),
	)
	require.NoError(t, err)
	assert.Equal(t, uint(1), pageDeleted.Total)

	// Filter all
	pageAll, err := sessRepo.GetListForUser(ctx, userID,
		session.WithLimit(10),
		session.WithDeleted(session.DeletedFilterAll),
	)
	require.NoError(t, err)
	assert.Equal(t, uint(2), pageAll.Total)
}

func TestSessionRepo_GetListForUser_Pagination(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)

	for i := 0; i < 5; i++ {
		_, err := sessRepo.Create(ctx, uuid.New(), userID, uuid.New().String())
		require.NoError(t, err)
	}

	page, err := sessRepo.GetListForUser(ctx, userID,
		session.WithLimit(2),
		session.WithOffset(0),
	)
	require.NoError(t, err)
	assert.Equal(t, uint(5), page.Total)
	assert.Len(t, page.Data, 2)

	page2, err := sessRepo.GetListForUser(ctx, userID,
		session.WithLimit(2),
		session.WithOffset(2),
	)
	require.NoError(t, err)
	assert.Len(t, page2.Data, 2)
}

func TestSessionRepo_GetToken(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)
	sessionID := uuid.New()

	_, err := sessRepo.Create(ctx, sessionID, userID, "myhash")
	require.NoError(t, err)

	hash, err := sessRepo.GetToken(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, "myhash", hash)
}

func TestSessionRepo_UpdateToken(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)
	sessionID := uuid.New()

	_, err := sessRepo.Create(ctx, sessionID, userID, "oldhash")
	require.NoError(t, err)

	updated, err := sessRepo.UpdateToken(ctx, sessionID, "newhash")
	require.NoError(t, err)
	assert.Equal(t, sessionID, updated.ID)

	hash, err := sessRepo.GetToken(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, "newhash", hash)
}

func TestSessionRepo_Delete(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)
	sessionID := uuid.New()

	_, err := sessRepo.Create(ctx, sessionID, userID, "hash")
	require.NoError(t, err)

	err = sessRepo.Delete(ctx, sessionID)
	require.NoError(t, err)

	// Session still accessible by GetByID (no deleted_at filter there)
	got, err := sessRepo.GetByID(ctx, sessionID)
	require.NoError(t, err)
	assert.NotNil(t, got.DeletedAt)
}

func TestSessionRepo_Delete_NotFound(t *testing.T) {
	_, sessRepo := newSessionRepos(t)

	err := sessRepo.Delete(context.Background(), uuid.New())
	assert.ErrorIs(t, err, errx.ErrorSessionNotFound)
}

func TestSessionRepo_DeleteOneForUser(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)
	sessionID := uuid.New()

	_, err := sessRepo.Create(ctx, sessionID, userID, "hash")
	require.NoError(t, err)

	err = sessRepo.DeleteOneForUser(ctx, userID, sessionID)
	require.NoError(t, err)

	got, err := sessRepo.GetByID(ctx, sessionID)
	require.NoError(t, err)
	assert.NotNil(t, got.DeletedAt)
}

func TestSessionRepo_DeleteOneForUser_WrongUser(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	acc1 := createUserForSession(t, accRepo)
	acc2 := createUserForSession(t, accRepo)
	sessionID := uuid.New()

	_, err := sessRepo.Create(ctx, sessionID, acc1, "hash")
	require.NoError(t, err)

	err = sessRepo.DeleteOneForUser(ctx, acc2, sessionID)
	assert.ErrorIs(t, err, errx.ErrorSessionNotFound)
}

func TestSessionRepo_DeleteManyForUser(t *testing.T) {
	accRepo, sessRepo := newSessionRepos(t)
	ctx := context.Background()

	userID := createUserForSession(t, accRepo)

	for i := 0; i < 3; i++ {
		_, err := sessRepo.Create(ctx, uuid.New(), userID, uuid.New().String())
		require.NoError(t, err)
	}

	ids, err := sessRepo.DeleteManyForUser(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, ids, 3)

	// All should be deleted now
	page, err := sessRepo.GetListForUser(ctx, userID,
		session.WithDeleted(session.DeletedFilterActive),
		session.WithLimit(10),
	)
	require.NoError(t, err)
	assert.Equal(t, uint(0), page.Total)
}
