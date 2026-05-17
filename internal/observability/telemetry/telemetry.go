package telemetry

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Config struct {
	// ServiceName прикрепляется ко всем трейсам и метрикам — по нему фильтруем в Grafana
	ServiceName string
	// CollectorEndpoint — адрес OTEL Collector, не Tempo напрямую
	// локально: "http://localhost:4318", в docker: "http://otel-collector:4318"
	CollectorEndpoint string
	// SamplingRatio — доля запросов которые трейсим: 1.0 = 100%, 0.1 = 10%
	SamplingRatio float64
}

// Init поднимает TracerProvider (трейсы) и MeterProvider (метрики),
// регистрирует их глобально. otelhttp и otelgrpc подхватят их автоматически.
// Возвращает shutdown — вызвать при остановке приложения для flush буферов.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	// Resource — метаданные "кто я", прикрепляются ко всей телеметрии.
	// resource.Default() добавляет OS, Go runtime, process info автоматически.
	// Мержим с нашим service.name поверх дефолтов.
	collectorURL, err := url.Parse(cfg.CollectorEndpoint)
	if err != nil {
		return nil, fmt.Errorf("parse collector endpoint: %w", err)
	}
	collectorHost := collectorURL.Host
	insecure := collectorURL.Scheme == "http"

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"", // пустой schema URL — иначе конфликт с resource.Default() который использует свою версию схемы
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	// — трейсы —

	// Экспортер отправляет spans в OTEL Collector по OTLP/HTTP протоколу.
	// Collector сам разберётся куда их дальше слать (Tempo, Jaeger, и т.д.).
	traceOpts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(collectorHost)}
	if insecure {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
	}
	traceExporter, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	traceProvider := sdktrace.NewTracerProvider(
		// WithBatcher — буферизует spans и отправляет пачками, не по одному.
		// Снижает нагрузку на сеть. Альтернатива: WithSyncer — отправляет сразу (только для тестов).
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		// TraceIDRatioBased — трейсим случайные X% запросов.
		// 1.0 в dev (всё), 0.1 в продакшне (каждый 10-й).
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRatio)),
	)

	// — метрики —

	// OTLP метрики уходят в Collector на тот же endpoint что и трейсы (/v1/metrics).
	// Collector экспортирует их в Prometheus формате на порту 8889 — Prometheus скрейпит Collector,
	// не приложение напрямую. Приложение знает только про один endpoint — коллектор.
	metricOpts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(collectorHost)}
	if insecure {
		metricOpts = append(metricOpts, otlpmetrichttp.WithInsecure())
	}
	metricExporter, err := otlpmetrichttp.New(ctx, metricOpts...)
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(30*time.Second),
		)),
		sdkmetric.WithResource(res),
	)

	// — глобальные провайдеры —

	// Регистрируем глобально — otelhttp middleware и otelgrpc interceptors
	// подхватят их автоматически через otel.GetTracerProvider() / otel.GetMeterProvider().
	otel.SetTracerProvider(traceProvider)
	otel.SetMeterProvider(meterProvider)

	// — Go runtime метрики —

	// runtime.Start запускает фоновую горутину, которая через Go runtime/metrics API
	// собирает goroutines count, heap alloc/inuse, GC pause duration и т.д.
	// Использует глобальный MeterProvider — поэтому вызываем ПОСЛЕ SetMeterProvider.
	// MinimumReadMemStatsInterval: не читать runtime.ReadMemStats чаще чем раз в секунду
	// (ReadMemStats — STW операция, блокирует все горутины на ~100мкс).
	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
		return nil, fmt.Errorf("start runtime instrumentation: %w", err)
	}

	// TextMapPropagator — как trace context передаётся между сервисами через HTTP заголовки.
	// TraceContext — стандарт W3C: заголовок "traceparent: 00-traceId-spanId-flags".
	// Baggage — дополнительные key-value метаданные в заголовке "baggage:".
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		// Shutdown флашит все буферы — важно вызвать при graceful shutdown,
		// иначе последние spans потеряются (они ещё в BatchSpanProcessor).
		if err := traceProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown trace provider: %w", err)
		}
		return meterProvider.Shutdown(ctx)
	}, nil
}
