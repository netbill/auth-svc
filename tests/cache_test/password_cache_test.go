package cache_test

import (
	"context"
	"testing"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/tests/testutil"
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

	err := cache.SetByID(ctx, pass)
	require.NoError(t, err)

	got, err := cache.GetByID(ctx, pass.AccountID)
	require.NoError(t, err)
	assert.Equal(t, pass.AccountID, got.AccountID)
	assert.Equal(t, "bcrypthash", got.Hash)
}

func TestPasswordCache_GetByID_Miss(t *testing.T) {
	setupCacheTest(t)
	cache := newPasswordCache(t)
	ctx := context.Background()

	_, err := cache.GetByID(ctx, testutil.RandomUUID())
	assert.ErrorIs(t, err, errx.ErrCacheMiss)
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

	err := cache.SetByID(ctx, pass)
	require.NoError(t, err)

	err = cache.DeleteByID(ctx, pass.AccountID)
	require.NoError(t, err)

	_, err = cache.GetByID(ctx, pass.AccountID)
	assert.ErrorIs(t, err, errx.ErrCacheMiss)
}

func TestPasswordCache_Overwrite(t *testing.T) {
	setupCacheTest(t)
	cache := newPasswordCache(t)
	ctx := context.Background()

	id := testutil.RandomUUID()

	pass := models.AccountPassword{AccountID: id, Hash: "oldhash", Version: 1}
	require.NoError(t, cache.SetByID(ctx, pass))

	pass.Hash = "newhash"
	pass.Version = 2
	require.NoError(t, cache.SetByID(ctx, pass))

	got, err := cache.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "newhash", got.Hash)
	assert.Equal(t, int32(2), got.Version)
}
