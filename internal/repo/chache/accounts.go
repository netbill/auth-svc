package chache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/redis/go-redis/v9"
)

type accountCacheMetrics interface {
	AccountCacheOp(ctx context.Context, err *error)
}

type AccountCache struct {
	client  *redis.Client
	ttl     time.Duration
	metrics accountCacheMetrics
	log     *log.Logger
}

func NewAccountCache(client *redis.Client, ttl time.Duration, m accountCacheMetrics, log *log.Logger) *AccountCache {
	return &AccountCache{client: client, ttl: ttl, metrics: m, log: log}
}

func accountKey(id uuid.UUID) string {
	return fmt.Sprintf("account:%s", id)
}

func (c *AccountCache) Set(ctx context.Context, account models.Account) error {
	if err := c.client.JSONSet(ctx, accountKey(account.ID), "$", account).Err(); err != nil {
		c.log.WithError(err).Error("account cache set failed", "account_id", account.ID)
		return err
	}
	if err := c.client.Expire(ctx, accountKey(account.ID), c.ttl).Err(); err != nil {
		c.log.WithError(err).Error("account cache expire failed", "account_id", account.ID)
		return err
	}
	return nil
}

func (c *AccountCache) Get(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	var err error
	defer c.metrics.AccountCacheOp(ctx, &err)

	val, err := c.client.JSONGet(ctx, accountKey(accountID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.Account{}, err
	case err != nil:
		c.log.WithError(err).Error("account cache get failed", "account_id", accountID)
		return models.Account{}, err
	}

	var account models.Account
	if err = json.Unmarshal([]byte(val), &account); err != nil {
		c.log.WithError(err).Error("account cache unmarshal failed", "account_id", accountID)
		return models.Account{}, err
	}

	return account, nil
}

func (c *AccountCache) Delete(ctx context.Context, accountID uuid.UUID) error {
	if err := c.client.Del(ctx, accountKey(accountID)).Err(); err != nil {
		c.log.WithError(err).Error("account cache delete failed", "account_id", accountID)
		return err
	}
	return nil
}
