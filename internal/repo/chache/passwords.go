package chache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/redis/go-redis/v9"
)

type PasswordCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewPasswordCache(
	client *redis.Client,
	ttl time.Duration,
) *PasswordCache {
	return &PasswordCache{
		client: client,
		ttl:    ttl,
	}
}

func passwordKey(accountID uuid.UUID) string {
	return fmt.Sprintf("account:password:%s", accountID)
}

func (c *PasswordCache) SetByID(ctx context.Context, password models.AccountPassword) error {
	if err := c.client.JSONSet(ctx, passwordKey(password.AccountID), "$", password).Err(); err != nil {
		return err
	}
	return c.client.Expire(ctx, passwordKey(password.AccountID), c.ttl).Err()
}

func (c *PasswordCache) GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error) {
	val, err := c.client.JSONGet(ctx, passwordKey(accountID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.AccountPassword{}, errx.ErrCacheMiss
	case err != nil:
		return models.AccountPassword{}, err
	}

	var password models.AccountPassword
	if err = json.Unmarshal([]byte(val), &password); err != nil {
		return models.AccountPassword{}, err
	}
	return password, nil
}

func (c *PasswordCache) DeleteByID(ctx context.Context, accountID uuid.UUID) error {
	return c.client.Del(ctx, passwordKey(accountID)).Err()
}
