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

func NewSessionCache(client *redis.Client, ttl time.Duration) *SessionCache {
	return &SessionCache{client: client, ttl: ttl}
}

func sessionKey(id uuid.UUID) string {
	return fmt.Sprintf("session:%s", id)
}

func (c *SessionCache) Set(ctx context.Context, session models.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, sessionKey(session.ID), data, c.ttl).Err()
}

func (c *SessionCache) GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error) {
	data, err := c.client.Get(ctx, sessionKey(sessionID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return models.Session{}, errx.ErrCacheMiss
		}
		return models.Session{}, err
	}

	var session models.Session
	if err = json.Unmarshal(data, &session); err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (c *SessionCache) DeleteByID(ctx context.Context, sessionID uuid.UUID) error {
	return c.client.Del(ctx, sessionKey(sessionID)).Err()
}
