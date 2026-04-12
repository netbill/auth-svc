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

type SessionCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewSessionCache(
	client *redis.Client,
	ttl time.Duration,
) *SessionCache {
	return &SessionCache{
		client: client,
		ttl:    ttl,
	}
}

func sessionKey(id uuid.UUID) string {
	return fmt.Sprintf("session:%s", id)
}

func (c *SessionCache) Set(ctx context.Context, session models.Session) error {
	if err := c.client.JSONSet(ctx, sessionKey(session.ID), "$", session).Err(); err != nil {
		return err
	}
	return c.client.Expire(ctx, sessionKey(session.ID), c.ttl).Err()
}

func (c *SessionCache) GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error) {
	val, err := c.client.JSONGet(ctx, sessionKey(sessionID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.Session{}, errx.ErrCacheMiss
	case err != nil:
		return models.Session{}, err
	}

	var session models.Session
	if err = json.Unmarshal([]byte(val), &session); err != nil {
		return models.Session{}, err
	}
	return session, nil
}

func (c *SessionCache) DeleteByID(ctx context.Context, sessionID uuid.UUID) error {
	return c.client.Del(ctx, sessionKey(sessionID)).Err()
}
