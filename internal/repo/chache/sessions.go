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

type sessionCacheMetrics interface {
	SessionCacheOp(ctx context.Context, err *error)
}

type SessionCache struct {
	client  *redis.Client
	ttl     time.Duration
	metrics sessionCacheMetrics
	log     *log.Logger
}

func NewSessionCache(client *redis.Client, ttl time.Duration, m sessionCacheMetrics, log *log.Logger) *SessionCache {
	return &SessionCache{client: client, ttl: ttl, metrics: m, log: log}
}

func sessionKey(id uuid.UUID) string {
	return fmt.Sprintf("session:%s", id)
}

func (c *SessionCache) Set(ctx context.Context, session models.Session) error {
	if err := c.client.JSONSet(ctx, sessionKey(session.ID), "$", session).Err(); err != nil {
		c.log.WithError(err).Error("session cache set failed", "session_id", session.ID)
		return err
	}
	if err := c.client.Expire(ctx, sessionKey(session.ID), c.ttl).Err(); err != nil {
		c.log.WithError(err).Error("session cache expire failed", "session_id", session.ID)
		return err
	}
	return nil
}

func (c *SessionCache) Get(ctx context.Context, sessionID uuid.UUID) (models.Session, error) {
	var err error
	defer c.metrics.SessionCacheOp(ctx, &err)

	val, err := c.client.JSONGet(ctx, sessionKey(sessionID), ".").Result()
	switch {
	case errors.Is(err, redis.Nil):
		return models.Session{}, err
	case err != nil:
		c.log.WithError(err).Error("session cache get failed", "session_id", sessionID)
		return models.Session{}, err
	}

	var session models.Session
	if err = json.Unmarshal([]byte(val), &session); err != nil {
		c.log.WithError(err).Error("session cache unmarshal failed", "session_id", sessionID)
		return models.Session{}, err
	}

	return session, nil
}

func (c *SessionCache) Delete(ctx context.Context, sessionID uuid.UUID) error {
	if err := c.client.Del(ctx, sessionKey(sessionID)).Err(); err != nil {
		c.log.WithError(err).Error("session cache delete failed", "session_id", sessionID)
		return err
	}
	return nil
}
