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

func NewPasswordCache(client *redis.Client, ttl time.Duration, m passwordCacheMetrics, log *log.Logger) *PasswordCache {
	return &PasswordCache{client: client, ttl: ttl, metrics: m, log: log}
}

func passwordKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:password:%s", userID)
}

func (c *PasswordCache) Set(ctx context.Context, password models.UserPassword) error {
	if err := c.client.JSONSet(ctx, passwordKey(password.UserID), "$", password).Err(); err != nil {
		c.log.WithError(err).Error("password cache set failed", "user_id", password.UserID)
		return err
	}
	if err := c.client.Expire(ctx, passwordKey(password.UserID), c.ttl).Err(); err != nil {
		c.log.WithError(err).Error("password cache expire failed", "user_id", password.UserID)
		return err
	}
	return nil
}

func (c *PasswordCache) Get(ctx context.Context, userID uuid.UUID) (models.UserPassword, error) {
	var err error
	defer c.metrics.PasswordCacheOp(ctx, &err)

	val, err := c.client.JSONGet(ctx, passwordKey(userID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.UserPassword{}, err
	case err != nil:
		c.log.WithError(err).Error("password cache get failed", "user_id", userID)
		return models.UserPassword{}, err
	}

	var password models.UserPassword
	if err = json.Unmarshal([]byte(val), &password); err != nil {
		c.log.WithError(err).Error("password cache unmarshal failed", "user_id", userID)
		return models.UserPassword{}, err
	}

	return password, nil
}

func (c *PasswordCache) Delete(ctx context.Context, userID uuid.UUID) error {
	if err := c.client.Del(ctx, passwordKey(userID)).Err(); err != nil {
		c.log.WithError(err).Error("password cache delete failed", "user_id", userID)
		return err
	}
	return nil
}
