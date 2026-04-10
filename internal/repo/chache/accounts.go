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

type AccountCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewAccountCache(client *redis.Client, ttl time.Duration) *AccountCache {
	return &AccountCache{client: client, ttl: ttl}
}

func accountKey(id uuid.UUID) string {
	return fmt.Sprintf("account:%s", id)
}

func (c *AccountCache) Set(ctx context.Context, account models.Account) error {
	data, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, accountKey(account.ID), data, c.ttl).Err()
}

func (c *AccountCache) GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	data, err := c.client.Get(ctx, accountKey(accountID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return models.Account{}, errx.ErrCacheMiss
		}
		return models.Account{}, err
	}

	var account models.Account
	if err = json.Unmarshal(data, &account); err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (c *AccountCache) Delete(ctx context.Context, accountID uuid.UUID) error {
	return c.client.Del(ctx, accountKey(accountID)).Err()
}
