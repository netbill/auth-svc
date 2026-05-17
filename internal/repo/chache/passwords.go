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

type passwordCacheMetrics interface {
	PasswordCacheOp(ctx context.Context, err *error)
}

type PasswordCache struct {
	client  *redis.Client
	ttl     time.Duration
	metrics passwordCacheMetrics
	log     *log.Logger
}

func NewPasswordCache(
	client *redis.Client,
	ttl time.Duration,
	m passwordCacheMetrics,
	log *log.Logger,
) *PasswordCache {
	return &PasswordCache{
		client:  client,
		ttl:     ttl,
		metrics: m,
		log:     log,
	}
}

func passwordKey(accountID uuid.UUID) string {
	return fmt.Sprintf("account:password:%s", accountID)
}

func (c *PasswordCache) Set(ctx context.Context, password models.AccountPassword) {
	go func() {
		if err := c.client.JSONSet(ctx, passwordKey(password.AccountID), "$", password).Err(); err != nil {
			c.log.WithError(err).Error("password cache set failed", "account", password.AccountID)
		}
		if err := c.client.Expire(ctx, passwordKey(password.AccountID), c.ttl).Err(); err != nil {
			c.log.WithError(err).Error("password cache expire failed", "account", password.AccountID)
		}
	}()
}

func (c *PasswordCache) Get(ctx context.Context, accountID uuid.UUID) (password models.AccountPassword, ok bool) {
	var err error
	defer c.metrics.PasswordCacheOp(ctx, &err)

	val, err := c.client.JSONGet(ctx, passwordKey(accountID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.AccountPassword{}, false
	case err != nil:
		c.log.Error("password cache get failed", "account_id", accountID, "error", err)
		return models.AccountPassword{}, false
	}

	if err = json.Unmarshal([]byte(val), &password); err != nil {
		c.log.Error("password cache unmarshal failed", "account_id", accountID, "error", err)
		return models.AccountPassword{}, false
	}

	return password, true
}

func (c *PasswordCache) Delete(ctx context.Context, accountID uuid.UUID) {
	go func() {
		if err := c.client.Del(ctx, passwordKey(accountID)).Err(); err != nil {
			c.log.WithError(err).Error("password cache delete failed", "account", accountID)
		}
	}()
}
