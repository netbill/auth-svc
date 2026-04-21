package chache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/redis/go-redis/v9"
)

type QRCache struct {
	client *redis.Client
}

func NewQRCache(client *redis.Client) *QRCache {
	return &QRCache{client: client}
}

func qrKey(token string) string {
	return fmt.Sprintf("qr:%s", token)
}

func (c *QRCache) Set(ctx context.Context, token string, status string, ttl time.Duration) error {
	if err := c.client.Set(ctx, qrKey(token), status, ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (c *QRCache) Get(ctx context.Context, token string) (string, error) {
	val, err := c.client.Get(ctx, qrKey(token)).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return "", errx.ErrCacheMiss
	case err != nil:
		return "", err
	}
	return val, nil
}
