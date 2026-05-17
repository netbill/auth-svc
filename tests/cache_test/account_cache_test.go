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

func TestAccountCache_SetAndGet(t *testing.T) {
	setupCacheTest(t)
	cache := newAccountCache(t)
	ctx := context.Background()

	acc := models.Account{
		ID:        testutil.RandomUUID(),
		Username:  "alice",
		Role:      "user",
		Version:   1,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	err := cache.Set(ctx, acc)
	require.NoError(t, err)

	got, err := cache.Get(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, acc.ID, got.ID)
	assert.Equal(t, acc.Username, got.Username)
	assert.Equal(t, acc.Role, got.Role)
}

func TestAccountCache_GetByID_Miss(t *testing.T) {
	setupCacheTest(t)
	cache := newAccountCache(t)
	ctx := context.Background()

	_, err := cache.Get(ctx, testutil.RandomUUID())
	assert.ErrorIs(t, err, redis.Nil)
}

func TestAccountCache_Delete(t *testing.T) {
	setupCacheTest(t)
	cache := newAccountCache(t)
	ctx := context.Background()

	acc := models.Account{
		ID:       testutil.RandomUUID(),
		Username: "bob",
		Role:     "user",
	}

	err := cache.Set(ctx, acc)
	require.NoError(t, err)

	err = cache.Delete(ctx, acc.ID)
	require.NoError(t, err)

	_, err = cache.Get(ctx, acc.ID)
	assert.ErrorIs(t, err, redis.Nil)
}

func TestAccountCache_Delete_NonExistent(t *testing.T) {
	setupCacheTest(t)
	cache := newAccountCache(t)
	ctx := context.Background()

	// Deleting a key that doesn't exist should not error
	err := cache.Delete(ctx, testutil.RandomUUID())
	assert.NoError(t, err)
}

func TestAccountCache_Overwrite(t *testing.T) {
	setupCacheTest(t)
	cache := newAccountCache(t)
	ctx := context.Background()

	acc := models.Account{
		ID:       testutil.RandomUUID(),
		Username: "charlie",
		Role:     "user",
		Version:  1,
	}

	err := cache.Set(ctx, acc)
	require.NoError(t, err)

	// Overwrite with updated version
	acc.Username = "charlie_updated"
	acc.Version = 2
	err = cache.Set(ctx, acc)
	require.NoError(t, err)

	got, err := cache.Get(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, "charlie_updated", got.Username)
	assert.Equal(t, int32(2), got.Version)
}

func TestAccountCache_MultipleAccounts(t *testing.T) {
	setupCacheTest(t)
	cache := newAccountCache(t)
	ctx := context.Background()

	acc1 := models.Account{ID: testutil.RandomUUID(), Username: "user1", Role: "user"}
	acc2 := models.Account{ID: testutil.RandomUUID(), Username: "user2", Role: "user"}

	require.NoError(t, cache.Set(ctx, acc1))
	require.NoError(t, cache.Set(ctx, acc2))

	got1, err := cache.Get(ctx, acc1.ID)
	require.NoError(t, err)
	assert.Equal(t, "user1", got1.Username)

	got2, err := cache.Get(ctx, acc2.ID)
	require.NoError(t, err)
	assert.Equal(t, "user2", got2.Username)
}
