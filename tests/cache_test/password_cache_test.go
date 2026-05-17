package cache_test

import (
	"context"
	"testing"

	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/tests/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordCache_SetAndGetByID(t *testing.T) {
	setupCacheTest(t)
	cache := newPasswordCache(t)
	ctx := context.Background()

	pass := models.AccountPassword{
		AccountID: testutil.RandomUUID(),
		Hash:      "bcrypthash",
		Version:   1,
	}

	err := cache.Set(ctx, pass)
	require.NoError(t, err)

	got, err := cache.Get(ctx, pass.AccountID)
	require.NoError(t, err)
	assert.Equal(t, pass.AccountID, got.AccountID)
	assert.Equal(t, "bcrypthash", got.Hash)
}

func TestPasswordCache_GetByID_Miss(t *testing.T) {
	setupCacheTest(t)
	cache := newPasswordCache(t)
	ctx := context.Background()

	_, err := cache.Get(ctx, testutil.RandomUUID())
	assert.ErrorIs(t, err, redis.Nil)
}

func TestPasswordCache_DeleteByID(t *testing.T) {
	setupCacheTest(t)
	cache := newPasswordCache(t)
	ctx := context.Background()

	pass := models.AccountPassword{
		AccountID: testutil.RandomUUID(),
		Hash:      "todelete",
		Version:   1,
	}

	err := cache.Set(ctx, pass)
	require.NoError(t, err)

	err = cache.Delete(ctx, pass.AccountID)
	require.NoError(t, err)

	_, err = cache.Get(ctx, pass.AccountID)
	assert.ErrorIs(t, err, redis.Nil)
}

func TestPasswordCache_Overwrite(t *testing.T) {
	setupCacheTest(t)
	cache := newPasswordCache(t)
	ctx := context.Background()

	id := testutil.RandomUUID()

	pass := models.AccountPassword{AccountID: id, Hash: "oldhash", Version: 1}
	require.NoError(t, cache.Set(ctx, pass))

	pass.Hash = "newhash"
	pass.Version = 2
	require.NoError(t, cache.Set(ctx, pass))

	got, err := cache.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "newhash", got.Hash)
	assert.Equal(t, int32(2), got.Version)
}
