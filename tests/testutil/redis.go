package testutil

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func NewRedisClient(t *testing.T, addr, password string, db int) *redis.Client {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	require.NoError(t, client.Ping(context.Background()).Err(), "connect to test redis")

	t.Cleanup(func() { _ = client.Close() })

	return client
}

// FlushRedis clears all keys in the test Redis DB.
func FlushRedis(t *testing.T, client *redis.Client) {
	t.Helper()
	require.NoError(t, client.FlushDB(context.Background()).Err(), "flush redis db")
}
