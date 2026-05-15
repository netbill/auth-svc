package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCache_SetAndGetByID(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	sess := models.Session{
		ID:        testutil.RandomUUID(),
		AccountID: testutil.RandomUUID(),
		Version:   1,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		LastUsed:  time.Now().UTC().Truncate(time.Second),
	}

	err := cache.Set(ctx, sess)
	require.NoError(t, err)

	got, err := cache.GetByID(ctx, sess.ID)
	require.NoError(t, err)
	assert.Equal(t, sess.ID, got.ID)
	assert.Equal(t, sess.AccountID, got.AccountID)
}

func TestSessionCache_GetByID_Miss(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	_, err := cache.GetByID(ctx, testutil.RandomUUID())
	assert.ErrorIs(t, err, errx.ErrCacheMiss)
}

func TestSessionCache_DeleteByID(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	sess := models.Session{
		ID:        testutil.RandomUUID(),
		AccountID: testutil.RandomUUID(),
		Version:   1,
	}

	err := cache.Set(ctx, sess)
	require.NoError(t, err)

	err = cache.DeleteByID(ctx, sess.ID)
	require.NoError(t, err)

	_, err = cache.GetByID(ctx, sess.ID)
	assert.ErrorIs(t, err, errx.ErrCacheMiss)
}

func TestSessionCache_DeleteByID_NonExistent(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	// Deleting a non-existent key should not error
	err := cache.DeleteByID(ctx, testutil.RandomUUID())
	assert.NoError(t, err)
}

func TestSessionCache_MultipleSessions(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	accountID := testutil.RandomUUID()
	sess1 := models.Session{ID: testutil.RandomUUID(), AccountID: accountID, Version: 1}
	sess2 := models.Session{ID: testutil.RandomUUID(), AccountID: accountID, Version: 1}

	require.NoError(t, cache.Set(ctx, sess1))
	require.NoError(t, cache.Set(ctx, sess2))

	got1, err := cache.GetByID(ctx, sess1.ID)
	require.NoError(t, err)
	assert.Equal(t, accountID, got1.AccountID)

	got2, err := cache.GetByID(ctx, sess2.ID)
	require.NoError(t, err)
	assert.Equal(t, accountID, got2.AccountID)

	// Delete one — the other should survive
	require.NoError(t, cache.DeleteByID(ctx, sess1.ID))

	_, err = cache.GetByID(ctx, sess1.ID)
	assert.ErrorIs(t, err, errx.ErrCacheMiss)

	_, err = cache.GetByID(ctx, sess2.ID)
	assert.NoError(t, err)
}
