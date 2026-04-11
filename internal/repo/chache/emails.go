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

type EmailCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewEmailCache(
	client *redis.Client,
	ttl time.Duration,
) *EmailCache {
	return &EmailCache{
		client: client,
		ttl:    ttl,
	}
}

func emailByIDKey(accountID uuid.UUID) string {
	return fmt.Sprintf("account:email:id:%s", accountID)
}

func emailByEmailKey(email string) string {
	return fmt.Sprintf("account:email:%s", email)
}

func (c *EmailCache) Set(ctx context.Context, email models.AccountEmail) error {
	data, err := json.Marshal(email)
	if err != nil {
		return err
	}

	if err = c.client.Set(ctx, emailByIDKey(email.AccountID), data, c.ttl).Err(); err != nil {
		return err
	}

	return c.client.Set(ctx, emailByEmailKey(email.Email), data, c.ttl).Err()
}

func (c *EmailCache) GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error) {
	data, err := c.client.Get(ctx, emailByIDKey(accountID)).Bytes()
	switch {
	case errors.Is(err, redis.Nil):
		return models.AccountEmail{}, errx.ErrCacheMiss
	case err != nil:
		return models.AccountEmail{}, err
	}

	var email models.AccountEmail
	if err = json.Unmarshal(data, &email); err != nil {
		return models.AccountEmail{}, err
	}

	return email, nil
}

func (c *EmailCache) GetByEmail(ctx context.Context, email string) (models.AccountEmail, error) {
	data, err := c.client.Get(ctx, emailByEmailKey(email)).Bytes()
	switch {
	case errors.Is(err, redis.Nil):
		return models.AccountEmail{}, errx.ErrCacheMiss
	case err != nil:
		return models.AccountEmail{}, err
	}

	var e models.AccountEmail
	if err = json.Unmarshal(data, &e); err != nil {
		return models.AccountEmail{}, err
	}

	return e, nil
}

func (c *EmailCache) DeleteByID(ctx context.Context, accountID uuid.UUID) error {
	return c.client.Del(ctx, emailByIDKey(accountID)).Err()
}

func (c *EmailCache) DeleteByEmail(ctx context.Context, email string) error {
	return c.client.Del(ctx, emailByEmailKey(email)).Err()
}
