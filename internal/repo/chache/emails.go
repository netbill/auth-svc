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

type emailCacheMetrics interface {
	EmailCacheOp(ctx context.Context, err *error)
}

type EmailCache struct {
	client  *redis.Client
	ttl     time.Duration
	metrics emailCacheMetrics
	log     *log.Logger
}

func NewEmailCache(client *redis.Client, ttl time.Duration, m emailCacheMetrics, log *log.Logger) *EmailCache {
	return &EmailCache{client: client, ttl: ttl, metrics: m, log: log}
}

func emailByIDKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:email:id:%s", userID)
}

func emailByEmailKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

func (c *EmailCache) Set(ctx context.Context, email models.UserEmail) error {
	if err := c.client.JSONSet(ctx, emailByIDKey(email.UserID), "$", email).Err(); err != nil {
		c.log.WithError(err).Error("email cache set by id failed")
		return err
	}
	if err := c.client.Expire(ctx, emailByIDKey(email.UserID), c.ttl).Err(); err != nil {
		c.log.WithError(err).Error("email cache expire by id failed")
		return err
	}
	if err := c.client.JSONSet(ctx, emailByEmailKey(email.Email), "$", email).Err(); err != nil {
		c.log.WithError(err).Error("email cache set by email failed")
		return err
	}
	if err := c.client.Expire(ctx, emailByEmailKey(email.Email), c.ttl).Err(); err != nil {
		c.log.WithError(err).Error("email cache expire by email failed")
		return err
	}
	return nil
}

func (c *EmailCache) GetByID(ctx context.Context, userID uuid.UUID) (models.UserEmail, error) {
	var err error
	defer c.metrics.EmailCacheOp(ctx, &err)

	val, err := c.client.JSONGet(ctx, emailByIDKey(userID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.UserEmail{}, err
	case err != nil:
		c.log.WithError(err).Error("email cache get by id failed", "user_id", userID)
		return models.UserEmail{}, err
	}

	var email models.UserEmail
	if err = json.Unmarshal([]byte(val), &email); err != nil {
		c.log.WithError(err).Error("email cache unmarshal failed", "user_id", userID)
		return models.UserEmail{}, err
	}

	return email, nil
}

func (c *EmailCache) GetByEmail(ctx context.Context, emailAddr string) (models.UserEmail, error) {
	val, err := c.client.JSONGet(ctx, emailByEmailKey(emailAddr), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.UserEmail{}, redis.Nil
	case err != nil:
		c.log.WithError(err).Error("email cache get by email failed", "email", emailAddr)
		return models.UserEmail{}, err
	}

	var email models.UserEmail
	if err = json.Unmarshal([]byte(val), &email); err != nil {
		c.log.WithError(err).Error("email cache unmarshal failed", "email", emailAddr)
		return models.UserEmail{}, err
	}

	return email, nil
}

func (c *EmailCache) DeleteByID(ctx context.Context, userID uuid.UUID) error {
	if err := c.client.Del(ctx, emailByIDKey(userID)).Err(); err != nil {
		c.log.WithError(err).Error("email cache delete by id failed", "user_id", userID)
		return err
	}
	return nil
}

func (c *EmailCache) DeleteByEmail(ctx context.Context, email string) error {
	if err := c.client.Del(ctx, emailByEmailKey(email)).Err(); err != nil {
		c.log.WithError(err).Error("email cache delete by email failed", "email", email)
		return err
	}
	return nil
}
