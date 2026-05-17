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

func NewAccountCache(
	client *redis.Client,
	ttl time.Duration,
	m accountCacheMetrics,
	log *log.Logger,
) *AccountCache {
	return &AccountCache{
		client:  client,
		ttl:     ttl,
		metrics: m,
		log:     log,
	}
}

func accountKey(id uuid.UUID) string {
	return fmt.Sprintf("account:%s", id)
}

func (c *AccountCache) Set(ctx context.Context, account models.Account) {
	go func() {
		if err := c.client.JSONSet(ctx, accountKey(account.ID), "$", account).Err(); err != nil {
			c.log.WithError(err).Error("failed to set account")
		}

		if err := c.client.Expire(ctx, accountKey(account.ID), c.ttl).Err(); err != nil {
			c.log.WithError(err).Error("failed to set account")
		}
	}()
}

func (c *AccountCache) Get(ctx context.Context, accountID uuid.UUID) (account models.Account, ok bool) {
	var err error
	defer c.metrics.AccountCacheOp(ctx, &err)

	val, err := c.client.JSONGet(ctx, accountKey(accountID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.Account{}, false
	case err != nil:
		c.log.WithError(err).Error("account cache get failed", "account_id", accountID)
		return models.Account{}, false
	}

	if err = json.Unmarshal([]byte(val), &account); err != nil {
		c.log.WithError(err).Error("account cache unmarshal failed", "account_id", accountID)
		return models.Account{}, false
	}

	return account, true
}

func (c *AccountCache) Delete(ctx context.Context, accountID uuid.UUID) {
	go func() {
		if err := c.client.Del(ctx, accountKey(accountID)).Err(); err != nil {
			c.log.WithError(err).Error("account cache delete failed", "account_id", accountID)
		}
	}()
}
