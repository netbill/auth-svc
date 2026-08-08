package metrics

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func statusFromErr(err *error) string {
	if err != nil && *err != nil {
		return "fail"
	}
	return "ok"
}

func (m *Metrics) UserCacheOp(ctx context.Context, err *error) {
	m.recordCacheOp(ctx, "user", err)
}

func (m *Metrics) SessionCacheOp(ctx context.Context, err *error) {
	m.recordCacheOp(ctx, "session", err)
}

func (m *Metrics) EmailCacheOp(ctx context.Context, err *error) {
	m.recordCacheOp(ctx, "email", err)
}

func (m *Metrics) PasswordCacheOp(ctx context.Context, err *error) {
	m.recordCacheOp(ctx, "password", err)
}

func (m *Metrics) recordCacheOp(ctx context.Context, entity string, err *error) {
	var result string
	switch {
	case *err == nil:
		result = "hit"
	case errors.Is(*err, redis.Nil):
		result = "miss"
	default:
		result = "error"
	}

	m.cacheOps.Add(ctx, 1, metric.WithAttributes(
		attribute.String("entity", entity),
		attribute.String("result", result),
	))
}
