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

func TestEmailCache_SetAndGetByID(t *testing.T) {
	setupCacheTest(t)
	cache := newEmailCache(t)
	ctx := context.Background()

	email := models.UserEmail{
		UserID:    testutil.RandomUUID(),
		Email:     "alice@example.com",
		Verified:  false,
		Version:   1,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	err := cache.Set(ctx, email)
	require.NoError(t, err)

	got, err := cache.GetByID(ctx, email.UserID)
	require.NoError(t, err)
	assert.Equal(t, email.UserID, got.UserID)
	assert.Equal(t, "alice@example.com", got.Email)
}

func TestEmailCache_SetAndGetByEmail(t *testing.T) {
	setupCacheTest(t)
	cache := newEmailCache(t)
	ctx := context.Background()

	email := models.UserEmail{
		UserID:   testutil.RandomUUID(),
		Email:    "bob@example.com",
		Verified: false,
		Version:  1,
	}

	err := cache.Set(ctx, email)
	require.NoError(t, err)

	got, err := cache.GetByEmail(ctx, "bob@example.com")
	require.NoError(t, err)
	assert.Equal(t, email.UserID, got.UserID)
	assert.Equal(t, "bob@example.com", got.Email)
}

func TestEmailCache_GetByID_Miss(t *testing.T) {
	setupCacheTest(t)
	cache := newEmailCache(t)
	ctx := context.Background()

	_, err := cache.GetByID(ctx, testutil.RandomUUID())
	assert.ErrorIs(t, err, redis.Nil)
}

func TestEmailCache_GetByEmail_Miss(t *testing.T) {
	setupCacheTest(t)
	cache := newEmailCache(t)
	ctx := context.Background()

	_, err := cache.GetByEmail(ctx, "nobody@example.com")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestEmailCache_DeleteByID(t *testing.T) {
	setupCacheTest(t)
	cache := newEmailCache(t)
	ctx := context.Background()

	email := models.UserEmail{
		UserID:  testutil.RandomUUID(),
		Email:   "charlie@example.com",
		Version: 1,
	}

	err := cache.Set(ctx, email)
	require.NoError(t, err)

	err = cache.DeleteByID(ctx, email.UserID)
	require.NoError(t, err)

	// ID-keyed entry should be gone
	_, err = cache.GetByID(ctx, email.UserID)
	assert.ErrorIs(t, err, redis.Nil)

	// Email-keyed entry should still exist (DeleteByID only removes the ID key)
	got, err := cache.GetByEmail(ctx, email.Email)
	require.NoError(t, err)
	assert.Equal(t, email.UserID, got.UserID)
}

func TestEmailCache_DeleteByEmail(t *testing.T) {
	setupCacheTest(t)
	cache := newEmailCache(t)
	ctx := context.Background()

	email := models.UserEmail{
		UserID:  testutil.RandomUUID(),
		Email:   "dave@example.com",
		Version: 1,
	}

	err := cache.Set(ctx, email)
	require.NoError(t, err)

	err = cache.DeleteByEmail(ctx, email.Email)
	require.NoError(t, err)

	_, err = cache.GetByEmail(ctx, email.Email)
	assert.ErrorIs(t, err, redis.Nil)

	// ID-keyed entry should still exist
	got, err := cache.GetByID(ctx, email.UserID)
	require.NoError(t, err)
	assert.Equal(t, email.Email, got.Email)
}

func TestEmailCache_SetStoresTwoKeys(t *testing.T) {
	setupCacheTest(t)
	cache := newEmailCache(t)
	ctx := context.Background()

	email := models.UserEmail{
		UserID:  testutil.RandomUUID(),
		Email:   "eve@example.com",
		Version: 1,
	}

	require.NoError(t, cache.Set(ctx, email))

	// Both lookups should work
	byID, err := cache.GetByID(ctx, email.UserID)
	require.NoError(t, err)
	assert.Equal(t, email.Email, byID.Email)

	byEmail, err := cache.GetByEmail(ctx, email.Email)
	require.NoError(t, err)
	assert.Equal(t, email.UserID, byEmail.UserID)
}
