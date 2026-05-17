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

func emailByIDKey(accountID uuid.UUID) string {
	return fmt.Sprintf("account:email:id:%s", accountID)
}

func emailByEmailKey(email string) string {
	return fmt.Sprintf("account:email:%s", email)
}

func (c *EmailCache) Set(ctx context.Context, email models.AccountEmail) error {
	if err := c.client.JSONSet(ctx, emailByIDKey(email.AccountID), "$", email).Err(); err != nil {
		c.log.WithError(err).Error("email cache set by id failed")
		return err
	}
	if err := c.client.Expire(ctx, emailByIDKey(email.AccountID), c.ttl).Err(); err != nil {
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

func (c *EmailCache) GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error) {
	var err error
	defer c.metrics.EmailCacheOp(ctx, &err)

	val, err := c.client.JSONGet(ctx, emailByIDKey(accountID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.AccountEmail{}, err
	case err != nil:
		c.log.WithError(err).Error("email cache get by id failed", "account_id", accountID)
		return models.AccountEmail{}, err
	}

	var email models.AccountEmail
	if err = json.Unmarshal([]byte(val), &email); err != nil {
		c.log.WithError(err).Error("email cache unmarshal failed", "account_id", accountID)
		return models.AccountEmail{}, err
	}

	return email, nil
}

func (c *EmailCache) GetByEmail(ctx context.Context, emailAddr string) (models.AccountEmail, error) {
	val, err := c.client.JSONGet(ctx, emailByEmailKey(emailAddr), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.AccountEmail{}, redis.Nil
	case err != nil:
		c.log.WithError(err).Error("email cache get by email failed", "email", emailAddr)
		return models.AccountEmail{}, err
	}

	var email models.AccountEmail
	if err = json.Unmarshal([]byte(val), &email); err != nil {
		c.log.WithError(err).Error("email cache unmarshal failed", "email", emailAddr)
		return models.AccountEmail{}, err
	}

	return email, nil
}

func (c *EmailCache) DeleteByID(ctx context.Context, accountID uuid.UUID) error {
	if err := c.client.Del(ctx, emailByIDKey(accountID)).Err(); err != nil {
		c.log.WithError(err).Error("email cache delete by id failed", "account_id", accountID)
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
