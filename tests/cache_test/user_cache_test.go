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

func TestUserCache_SetAndGet(t *testing.T) {
	setupCacheTest(t)
	cache := newUserCache(t)
	ctx := context.Background()

	acc := models.User{
		ID:        testutil.RandomUUID(),
		Role:      "user",
		Version:   1,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	err := cache.Set(ctx, acc)
	require.NoError(t, err)

	got, err := cache.Get(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, acc.ID, got.ID)
	assert.Equal(t, acc.Role, got.Role)
}

func TestUserCache_GetByID_Miss(t *testing.T) {
	setupCacheTest(t)
	cache := newUserCache(t)
	ctx := context.Background()

	_, err := cache.Get(ctx, testutil.RandomUUID())
	assert.ErrorIs(t, err, redis.Nil)
}

func TestUserCache_Delete(t *testing.T) {
	setupCacheTest(t)
	cache := newUserCache(t)
	ctx := context.Background()

	acc := models.User{
		ID:   testutil.RandomUUID(),
		Role: "user",
	}

	err := cache.Set(ctx, acc)
	require.NoError(t, err)

	err = cache.Delete(ctx, acc.ID)
	require.NoError(t, err)

	_, err = cache.Get(ctx, acc.ID)
	assert.ErrorIs(t, err, redis.Nil)
}

func TestUserCache_Delete_NonExistent(t *testing.T) {
	setupCacheTest(t)
	cache := newUserCache(t)
	ctx := context.Background()

	// Deleting a key that doesn't exist should not error
	err := cache.Delete(ctx, testutil.RandomUUID())
	assert.NoError(t, err)
}

func TestUserCache_Overwrite(t *testing.T) {
	setupCacheTest(t)
	cache := newUserCache(t)
	ctx := context.Background()

	acc := models.User{
		ID:      testutil.RandomUUID(),
		Role:    "user",
		Version: 1,
	}

	err := cache.Set(ctx, acc)
	require.NoError(t, err)

	// Overwrite with updated version
	acc.Role = "admin"
	acc.Version = 2
	err = cache.Set(ctx, acc)
	require.NoError(t, err)

	got, err := cache.Get(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", got.Role)
	assert.Equal(t, int32(2), got.Version)
}

func TestUserCache_MultipleUsers(t *testing.T) {
	setupCacheTest(t)
	cache := newUserCache(t)
	ctx := context.Background()

	acc1 := models.User{ID: testutil.RandomUUID(), Role: "user"}
	acc2 := models.User{ID: testutil.RandomUUID(), Role: "admin"}

	require.NoError(t, cache.Set(ctx, acc1))
	require.NoError(t, cache.Set(ctx, acc2))

	got1, err := cache.Get(ctx, acc1.ID)
	require.NoError(t, err)
	assert.Equal(t, "user", got1.Role)

	got2, err := cache.Get(ctx, acc2.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", got2.Role)
}
