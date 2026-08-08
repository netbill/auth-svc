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

type userCacheMetrics interface {
	UserCacheOp(ctx context.Context, err *error)
}

type UserCache struct {
	client  *redis.Client
	ttl     time.Duration
	metrics userCacheMetrics
	log     *log.Logger
}

func NewUserCache(client *redis.Client, ttl time.Duration, m userCacheMetrics, log *log.Logger) *UserCache {
	return &UserCache{client: client, ttl: ttl, metrics: m, log: log}
}

func userKey(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id)
}

func (c *UserCache) Set(ctx context.Context, user models.User) error {
	if err := c.client.JSONSet(ctx, userKey(user.ID), "$", user).Err(); err != nil {
		c.log.WithError(err).Error("user cache set failed", "user_id", user.ID)
		return err
	}
	if err := c.client.Expire(ctx, userKey(user.ID), c.ttl).Err(); err != nil {
		c.log.WithError(err).Error("user cache expire failed", "user_id", user.ID)
		return err
	}
	return nil
}

func (c *UserCache) Get(ctx context.Context, userID uuid.UUID) (models.User, error) {
	var err error
	defer c.metrics.UserCacheOp(ctx, &err)

	val, err := c.client.JSONGet(ctx, userKey(userID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.User{}, err
	case err != nil:
		c.log.WithError(err).Error("user cache get failed", "user_id", userID)
		return models.User{}, err
	}

	var user models.User
	if err = json.Unmarshal([]byte(val), &user); err != nil {
		c.log.WithError(err).Error("user cache unmarshal failed", "user_id", userID)
		return models.User{}, err
	}

	return user, nil
}

func (c *UserCache) Delete(ctx context.Context, userID uuid.UUID) error {
	if err := c.client.Del(ctx, userKey(userID)).Err(); err != nil {
		c.log.WithError(err).Error("user cache delete failed", "user_id", userID)
		return err
	}
	return nil
}
