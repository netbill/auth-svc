package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCache_SetAndGetByID(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	sess := models.Session{
		ID:        testutil.RandomUUID(),
		UserID:    testutil.RandomUUID(),
		Version:   1,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		LastUsed:  time.Now().UTC().Truncate(time.Second),
	}

	err := cache.Set(ctx, sess)
	require.NoError(t, err)

	got, err := cache.Get(ctx, sess.ID)
	require.NoError(t, err)
	assert.Equal(t, sess.ID, got.ID)
	assert.Equal(t, sess.UserID, got.UserID)
}

func TestSessionCache_GetByID_Miss(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	_, err := cache.Get(ctx, testutil.RandomUUID())
	assert.ErrorIs(t, err, redis.Nil)
}

func TestSessionCache_DeleteByID(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	sess := models.Session{
		ID:      testutil.RandomUUID(),
		UserID:  testutil.RandomUUID(),
		Version: 1,
	}

	err := cache.Set(ctx, sess)
	require.NoError(t, err)

	err = cache.Delete(ctx, sess.ID)
	require.NoError(t, err)

	_, err = cache.Get(ctx, sess.ID)
	assert.ErrorIs(t, err, redis.Nil)
}

func TestSessionCache_DeleteByID_NonExistent(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	// Deleting a non-existent key should not error
	err := cache.Delete(ctx, testutil.RandomUUID())
	assert.NoError(t, err)
}

func TestSessionCache_MultipleSessions(t *testing.T) {
	setupCacheTest(t)
	cache := newSessionCache(t)
	ctx := context.Background()

	userID := testutil.RandomUUID()
	sess1 := models.Session{ID: testutil.RandomUUID(), UserID: userID, Version: 1}
	sess2 := models.Session{ID: testutil.RandomUUID(), UserID: userID, Version: 1}

	require.NoError(t, cache.Set(ctx, sess1))
	require.NoError(t, cache.Set(ctx, sess2))

	got1, err := cache.Get(ctx, sess1.ID)
	require.NoError(t, err)
	assert.Equal(t, userID, got1.UserID)

	got2, err := cache.Get(ctx, sess2.ID)
	require.NoError(t, err)
	assert.Equal(t, userID, got2.UserID)

	// Delete one — the other should survive
	require.NoError(t, cache.Delete(ctx, sess1.ID))

	_, err = cache.Get(ctx, sess1.ID)
	assert.ErrorIs(t, err, redis.Nil)

	_, err = cache.Get(ctx, sess2.ID)
	assert.NoError(t, err)
}
