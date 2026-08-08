package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (m *Metrics) RecordEmailLogin(ctx context.Context, err *error) {
	m.logins.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", "email"),
		attribute.String("status", statusFromErr(err)),
	))
}

func (m *Metrics) RecordGoogleLogin(ctx context.Context, err *error) {
	m.logins.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", "google"),
		attribute.String("status", statusFromErr(err)),
	))
}

func (m *Metrics) RecordQRLogin(ctx context.Context, err *error) {
	m.logins.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", "qr"),
		attribute.String("status", statusFromErr(err)),
	))
}

func (m *Metrics) RecordRegistration(ctx context.Context, err *error) {
	m.registrations.Add(ctx, 1, metric.WithAttributes(
		attribute.String("status", statusFromErr(err)),
	))
}

func (m *Metrics) RecordSessionDeleted(ctx context.Context, scope string, err *error) {
	m.sessionDeletes.Add(ctx, 1, metric.WithAttributes(
		attribute.String("scope", scope),
		attribute.String("status", statusFromErr(err)),
	))
}

func (m *Metrics) RecordTokenRefresh(ctx context.Context, err *error) {
	m.tokenRefreshes.Add(ctx, 1, metric.WithAttributes(
		attribute.String("status", statusFromErr(err)),
	))
}
