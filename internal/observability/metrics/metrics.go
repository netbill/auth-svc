package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Metrics хранит все бизнес-счётчики сервиса.
// Создаётся один раз при старте, передаётся в сервисы как зависимость.
type Metrics struct {
	logins         metric.Int64Counter
	registrations  metric.Int64Counter
	sessionDeletes metric.Int64Counter
	tokenRefreshes metric.Int64Counter
	cacheOps       metric.Int64Counter
}

func New() (*Metrics, error) {
	// Meter берём из глобального MeterProvider — он зарегистрирован в telemetry.Init.
	// "auth-svc" — namespace, будет префиксом в Prometheus: auth_svc_auth_logins_total.
	meter := otel.GetMeterProvider().Meter("auth-svc")

	logins, err := meter.Int64Counter("auth.logins_total",
		metric.WithDescription("Login attempts by method (email|google|qr) and status (ok|fail)"),
	)
	if err != nil {
		return nil, fmt.Errorf("create logins counter: %w", err)
	}

	registrations, err := meter.Int64Counter("auth.registrations_total",
		metric.WithDescription("Registration attempts by status (ok|fail)"),
	)
	if err != nil {
		return nil, fmt.Errorf("create registrations counter: %w", err)
	}

	sessionDeletes, err := meter.Int64Counter("auth.sessions_deleted_total",
		metric.WithDescription("Sessions deleted by scope (single|all)"),
	)
	if err != nil {
		return nil, fmt.Errorf("create session_deletes counter: %w", err)
	}

	tokenRefreshes, err := meter.Int64Counter("auth.token_refreshes_total",
		metric.WithDescription("Token refresh attempts by status (ok|fail)"),
	)
	if err != nil {
		return nil, fmt.Errorf("create token_refreshes counter: %w", err)
	}

	cacheOps, err := meter.Int64Counter("auth.cache_operations_total",
		metric.WithDescription("Cache operations by entity (user|session|email|password) and result (hit|miss)"),
	)
	if err != nil {
		return nil, fmt.Errorf("create cache_ops counter: %w", err)
	}

	return &Metrics{
		logins:         logins,
		registrations:  registrations,
		sessionDeletes: sessionDeletes,
		tokenRefreshes: tokenRefreshes,
		cacheOps:       cacheOps,
	}, nil
}
