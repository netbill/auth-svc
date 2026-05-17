# auth-svc — Claude Context

## Проект

Go микросервис аутентификации (`github.com/netbill/auth-svc`).
Отвечает за регистрацию, логин, сессии, управление аккаунтами.

## Стек

- **Go 1.25**, chi router, pgx v5
- **PostgreSQL** — основная БД, миграции через `sql-migrate`
- **Redis** — кеш через `github.com/redis/go-redis/v9` (требует `go mod vendor`)
- **Kafka** — события через `github.com/netbill/eventbox` (outbox/inbox pattern)
- **Debezium** — подключён (ветка `feat/debezium`), читает WAL → Kafka через Outbox Event Router SMT
- **gRPC** — реализован (`internal/api/grpc/`), запускается параллельно с REST
- **gorilla/websocket v1.5.3** — QR-логин через WebSocket

## Структура

```
cmd/auth-svc/          — entrypoint
internal/
  app/                 — запуск приложения (run.go, migrations.go, events.go)
  api/
    rest/              — chi router, контроллеры, middleware
    grpc/              — gRPC сервер (auth, account, session)
    ws/                — WS сервис QR-логина; service.go + qr.go + responses/token.go
  modules/
    auth/              — ValidateSession (переиспользуется в account и session)
    account/           — регистрация, управление аккаунтом
    session/           — логин, сессии, токены, QR-логин; list.go — ListSessionsOption types
  messenger/
    consumer.go        — Kafka consumer (читает OrganizationsV1, OrgMembersV1)
    producer.go        — Kafka producer (пишет AccountsTopicV1)
    publisher/         — пишет события в outbox таблицу
    evcontroller/      — хендлеры входящих событий (org, org_member)
  bus/
    pub.go             — Publisher: низкоуровневый Publish в Redis Pub/Sub
    sub.go             — Subscriber: низкоуровневый Subscribe из Redis Pub/Sub → (<-chan []byte, cleanup)
    pool.go            — Broker: надстройка над pub+sub; PublishQRToken/SubscribeQRToken с префиксом ключа qr-token:{token}
  repo/
    pg/                — PostgreSQL реализации (accounts, emails, passwords, sessions, outbox)
    chache/            — Redis реализации (accounts, emails, passwords, sessions, qr)
  errx/                — доменные ошибки (ape)
  models/              — доменные модели
pkg/
  log/                 — логгер
  resources/           — OpenAPI генерированные модели и конфиг
```

## Текущая миграция (feat/debezium)

### Что уже сделано
- `outbox_events` таблица упрощена — убраны поля status, reserved_by, attempts (они нужны были только OutboxWorker)
- `docker-compose.yml` — добавлен `wal_level=logical` для PostgreSQL
- Debezium коннектор настроен в отдельном kafka-репо (один коннектор на всю инфраструктуру, dev окружение)

### Что удаляется
- `OutboxWorker` (`internal/messenger/outbox_worker.go`) — polling outbox_events → Kafka, заменён Debezium
- `InboxWorker` (`internal/messenger/inbox_worker.go`) — polling inbox_events → handlers, заменяется прямой обработкой в Consumer
- `inbox_events` таблица — больше не нужна (см. новую архитектуру inbox ниже)
- Cleanup функции outbox в `internal/app/events.go`
- Блок запуска outboxWorker и inboxWorker в `internal/app/run.go`

### Новая архитектура Inbox

**Принцип:** Consumer читает из Kafka и сразу вызывает хендлер без промежуточной таблицы.

```
Kafka partition N → Consumer → BEGIN → handler + ON CONFLICT DO NOTHING → COMMIT → commit offset
```

**Idempotency:** через уникальный ключ бизнес-сущности, не через inbox таблицу:
```sql
INSERT INTO organizations (id, ...) VALUES (...) ON CONFLICT (id) DO NOTHING
```

**Retry через "Listen to Yourself" pattern:**
- При ошибке хендлера — INSERT в outbox_events (retry) в той же DB транзакции
- Debezium читает WAL → пишет в retry-topic
- Consumer читает retry-topic → handler снова
- При превышении attempts → INSERT в outbox_events (dlq) → Debezium → dlq-topic

```
main-topic → handler OK  → commit offset
           → handler ERR → INSERT outbox(retry, attempts+1) → commit offset
                               ↓
                          Debezium → retry-topic
                               ↓
                          attempts >= max → outbox(dlq) → dlq-topic
```

**Гарантии доставки по типу события:**
- `OrgCreated`, `OrgDeleted`, `OrgMemberCreated`, `OrgMemberDeleted` — exactly-once (критично, порядок важен)
- fire-and-forget события (лайки, просмотры) — at-most-once, коммит сразу

**Retry backoff — решение принято:** отдельный топик на каждый уровень задержки. Конвенция именования: `{orig_topic}.retry.{delay}`, например:
```
organizations.retry.1m
organizations.retry.2m
organizations.retry.4m
organizations.retry.8m
organizations.retry.16m
organizations.retry.32m
organizations.retry.60m
organizations.dlq
```
- `attempts` и `retry_at` передаются в payload сообщения (в outbox_events записи нет после доставки Debezium)
- Ключ партиционирования `org_id` одинаковый на всех топиках — порядок сохраняется через retry
- Конкурентного чтения нет — Kafka consumer group распределяет партиции между инстансами автоматически

### Outbox (уже работает через Debezium)
- Publisher (`internal/messenger/publisher/`) пишет в outbox_events как раньше
- Debezium читает WAL и доставляет в Kafka — нет polling, нет reserved_by логики
- outbox_events таблица остаётся, но упрощена

## Архитектурный контекст (для понимания решений)

Проект позиционируется как масштабируемая система. При обсуждении архитектуры учитывать:

- **Шардирование БД** — несколько шардов, несколько инстансов сервиса
- **Ключ партиционирования** — `org_id` или аналог, детерминированно маппится на партицию Kafka и шард БД: `hash(key) % num_partitions`, `num_partitions` кратно `num_shards`
- **Координация между инстансами** — через Kafka consumer group + partition assignment, не через distributed locks
- **Geo-distribution** — концептуально несколько регионов, MirrorMaker 2 для репликации между Kafka кластерами
- **CAP trade-off** — для auth выбран CP (strong consistency важнее availability при сетевых сбоях)

## Кеш (Redis) — архитектурные решения

- **Библиотека:** TBD (go-redis или rueidis)
- **Паттерн:** cache-aside, горутина прогрева **после** коммита транзакции
- **Инвалидация:** в горутине после успешной мутации

### Что кешируется
- `account` — по ID (основной ключ)
- `accountEmail` — по ID, по email
- `session` — по ID

### Что НЕ кешируется / почему
- **Login lookup (email/username → account)** — CP при логине, кеш не используется
- **Списки сессий** (`GetListForAccount`) — не кешируются

### TTL стратегия — фиксированный TTL (решение принято)

**Решение:** фиксированный TTL, lazy TTL отклонён.

- **Lazy TTL отклонён** — горячий аккаунт/сессия читается постоянно → TTL бесконечно продлевается → данные могут быть неактуальны часами. От lazy TTL больше проблем чем пользы.
- **Фиксированный TTL** даёт предсказуемое окно неактуальности независимо от нагрузки — SLA на staleness.
- **Текущий код** уже реализует это корректно: `Set` вызывается только после DB read или DB write, при cache hit TTL не продлевается.

**Проблема инвалидации сессий (TOCTOU race):**
- `DeleteMySessions` удаляет из DB, но горутина параллельного запроса может записать сессию обратно в кеш уже после удаления
- Распределённые транзакции PG↔Redis невозможны, при шардировании (10 PG + 10 Redis) тем более
- **Принято:** eventual consistency для сессионного кеша — окно неактуальности ограничено TTL, это приемлемо

**Открытый вопрос:** конкретные значения TTL для сессий и аккаунта — не определены.

### Модули и кеш
- `auth.Service` — читает accountCache + sessionsCache для ValidateSession
- `account.Service` — читает/пишет accountCache, emailCache, passwordCache; при DeleteMyAccount инвалидирует sessionsCache
- `session.Service` — пишет accountCache + sessionsCache после логина; читает для Refresh

### Почему не вынесли кеш в отдельный слой
Обсуждали CachingRepo абстракцию — отказались:
- Service превратился бы в тонкую обёртку
- Горутины внутри транзакции небезопасны (race: транзакция откатилась, кеш уже записан)
- Кеш остаётся в Service, горутины только после `tx.Transaction()` вернул `nil`

### Почему не используем Debezium для инвалидации кеша
- WAL не содержит SELECT — cache warming через события невозможен без костылей
- Для инвалидации добавляет latency (WAL → Debezium → Kafka → Consumer → Redis)
- Горутина после транзакции — проще и достаточно

## Coding patterns

### pg репозитории
- **Константы таблиц/колонок** — на уровне пакета: `const (accountsTable = "accounts"; accountsCols = "id, ...")`
- **SQL запросы** — `const query = ...` внутри каждого метода, используют конкатенацию с пакетными константами
- **scan хелперы** — `switch { case ErrNoRows → domain error; case err != nil → wrap }`. Проверка `deleted_at != nil` в scan хелперах **не используется** — фильтрация через SQL (`AND deleted_at IS NULL`). Исключение: методы с опцией `WithDeleted` строят запрос динамически
- **Методы с `RETURNING`** — Delete методы возвращают `(Entity, error)` через `RETURNING cols` + inline scan (не через scan хелпер, т.к. хелпер может фильтровать deleted). Используется для получения актуального состояния (deleted_at, version) для outbox event
- **Unique constraint** — `pgconn.PgError` с кодом `23505` → доменная ошибка (`ErrorUsernameAlreadyTaken`, `ErrorEmailAlreadyExist`)
- **Без query builders** — только чистый SQL с плейсхолдерами `$1, $2, ...`
- **Ошибки** — всегда `.Raise(fmt.Errorf("context: %v", id))`, никогда `.Raise(nil)` или голый `errx.ErrorXxx`

### cache репозитории
- **Ошибка Redis** — `switch { case errors.Is(err, redis.Nil): → ErrCacheMiss; case err != nil: → err }`
- **JSON сериализация** — RedisJSON (`JSONSet`/`JSONGet` напрямую на `*redis.Client`, не через `client.JSON()`)

### Option types
- `ListSessionsOption` и типы — в domain пакете (`internal/modules/session/list.go`)
- `GetAccountOption` / `DeletedFilter` — в `internal/modules/account/options.go`, используется для `emailRepo.GetByID(ctx, id, WithDeleted(DeletedFilterAll))`
- pg пакет импортирует domain пакет для типов и строит SQL условия инлайн через `switch opts.Deleted`

### Триггеры БД (migrations/schema/002_account.sql)
- `cascade_account_soft_delete` — AFTER UPDATE ON accounts: при soft-delete аккаунта автоматически ставит `deleted_at` на `account_emails` и `account_passwords`
- `forbid_manual_soft_delete_account_email` — запрещает ручной soft-delete email пока аккаунт активен (`pg_trigger_depth() = 0`)
- **Следствие для `DeleteMyAccount`**: нельзя вызывать `emailRepo.Delete` ни до, ни после `accountRepo.Delete` — триггер делает каскад сам. Email получаем через `GetByID(..., WithDeleted(DeletedFilterAll))` уже после каскада. Сессии триггер **не трогает** — их удаляет Go-код через `sessionRepo.DeleteManyForAccount` внутри той же транзакции

## Зависимости eventbox

- `github.com/netbill/eventbox v0.1.14` — outbox/inbox interfaces + workers
- `github.com/netbill/evtypes v0.1.3` — топики и типы событий
- `github.com/netbill/pgdbx v0.3.1` — PostgreSQL helpers
- `github.com/redis/go-redis/v9` — Redis клиент (требует `go mod vendor`)

## Конфигурация

`config.yaml` / `config.example.yaml` — основной конфиг через viper.
`config_docker.yaml` — для docker окружения.

## QR-код логин (WebSocket) — реализовано

Фича "войти через QR-код" — как в Telegram Web. Десктоп открывает WS, получает токен, рендерит QR. Мобильный (уже залогинен) сканирует и делает POST /qr/confirm. Сервер создаёт сессию и пушит токены в WS десктопа.

### Флоу

```
Desktop  → GET /auth-svc/v1/login/qr  (WS upgrade, без auth)
Instance-1 → Broker.SubscribeQRToken("qr-token:{uuid}") → шлёт {"qr_token":"..."} по WS

Mobile   → POST /auth-svc/v1/login/qr/confirm  (bearer token, body: {"qr_token":"..."})
Instance-2 → session.ConfirmQRToken → контроллер → Broker.PublishQRToken("qr-token:{uuid}", tokensJSON)

Instance-1 → получает из Pub/Sub → шлёт payload по WS → закрывает
Mobile     → получает 204 No Content
```

### Архитектурные решения

- **Redis Pub/Sub** для координации между инстансами — WS stateful, confirm может попасть на другой инстанс
- **Subscribe до WriteJSON** — иначе race: мобильный успеет подтвердить до Subscribe
- **Горутина чтения** в WS handler — gorilla требует читать фреймы для детекции disconnect
- **Publish в контроллере** после `session.ConfirmQRToken` — бизнес логика не знает про Pub/Sub
- **Broker** (`internal/bus/pool.go`) — надстройка над Publisher+Subscriber; инкапсулирует префикс ключа `qr-token:{token}`; нет in-memory map, только Redis
- TTL pending: 2 мин, TTL confirmed: 30 сек

### Redis структура

- Ключ состояния: `qr:{uuid}`, значение: plain string `"pending"` | `"confirmed"` (в `chache/qr.go`)
- Pub/Sub канал: `qr-token:{uuid}`, payload: JSON `models.TokensPair`

### Файлы

```
internal/repo/chache/qr.go            ← QRCache: Set/Get (состояние в Redis key-value)
internal/bus/pub.go                   ← Publisher: низкоуровневый Publish
internal/bus/sub.go                   ← Subscriber: низкоуровневый Subscribe → <-chan []byte
internal/bus/pool.go                  ← Broker: PublishQRToken/SubscribeQRToken с префиксом ключа
internal/api/ws/service.go            ← ws.Service: upgrader + зависимости
internal/api/ws/qr.go                 ← LoginWithQR: весь WS lifecycle QR-логина
internal/api/ws/responses/token.go    ← WS response helpers: QRToken, AuthTokensPair, TokensExpiredError
internal/modules/session/qr.go        ← CreateQRToken, ConfirmQRToken (бизнес логика, без Pub/Sub)
internal/api/rest/requests/qr.go      ← парсинг confirm запроса
internal/api/rest/controller/qr.go    ← QRController: QRConnect (WS) + QRConfirm (REST + Publish)
```

Роуты:
- `GET /auth-svc/v1/login/qr` — WS upgrade, без auth
- `POST /auth-svc/v1/login/qr/confirm` — REST, bearer token → 204 No Content

Errx: `ErrorQRTokenNotFound`, `ErrorQRTokenAlreadyConfirmed` — в `errx/session.go`

---

## Observability — план реализации (Prometheus + Grafana + OTEL)

### Поток данных

```
auth-svc
  │
  ├─► OTLP HTTP :4318 ────────► OTEL Collector ──────────► Tempo ──────► Grafana
  │   (traces: каждый span)      (буферизует,               (хранит        (поиск
  │                               ретраит,                   трейсы)        трейсов)
  │                               фильтрует)
  │
  └─► GET /metrics ──────────► Prometheus ──────────────────────────────► Grafana
      (Prometheus scrape         (хранит метрики,                          (графики,
       раз в 15 сек)              time-series DB)                           алерты)
```

**Метрики** — агрегат: "за 5 минут 200 req, 3 ошибки, p99=45ms"
**Трейсы** — конкретный запрос: "POST /login = 120ms: 5ms parse + 80ms DB + 35ms Redis"

**OTEL Collector используем** — приложение знает только один endpoint (коллектор). Бэкенды (Tempo, Jaeger, любой другой) меняются только в конфиге коллектора без изменений Go кода. Коллектор буферизует spans и ретраит при временной недоступности Tempo.

### Архитектурные решения (принятые)

- Traces backend: **Tempo** (нативная интеграция с Grafana, одна экосистема)
- Traces routing: **OTEL Collector** — приложение шлёт OTLP в коллектор, коллектор шлёт в Tempo; смена бэкенда = правка конфига коллектора без изменений кода
- Metrics: **Prometheus scrape** — pull model, Prometheus сам приходит за метриками на `/metrics`
- Span propagation: **W3C TraceContext** заголовки (стандарт, поддерживается везде)
- Sampling: **100%** в dev (трейсим всё), в продакшне снизить до 10-20%
- OTEL SDK версия: **v1.x** (stable, не experimental)

### Новые файлы

```
internal/telemetry/
  telemetry.go     — Init(ctx, cfg) → (shutdown func, error): поднимает TracerProvider + MeterProvider
  config.go        — OTELConfig: ServiceName, CollectorEndpoint, SamplingRatio

observability/
  otelcol-config.yaml     — OTEL Collector: receivers(otlp) → exporters(otlphttp→tempo)
  prometheus.yml          — scrape_configs: auth-svc /metrics каждые 15 сек
  grafana/
    provisioning/
      datasources/
        datasources.yaml  — авто-подключение Prometheus + Tempo к Grafana
      dashboards/
        dashboards.yaml   — авто-загрузка dashboard JSON
    dashboards/
      auth-svc.json       — основной дашборд (HTTP rps, latency, errors, cache hits)
```

### Фаза 1 — Инфраструктура (docker-compose)

**Шаг 1.** Создать `observability/prometheus.yml` — конфиг Prometheus:
```yaml
scrape_configs:
  - job_name: auth-svc
    static_configs:
      - targets: ['auth-svc:8000']  # внутри docker сети
    metrics_path: /metrics
    scrape_interval: 15s
```

**Шаг 2.** Создать `observability/grafana/provisioning/datasources/datasources.yaml` — авто-подключение источников данных (Prometheus + Tempo) при старте Grafana.

**Шаг 3.** Создать `observability/otelcol-config.yaml` — конфиг OTEL Collector:
```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: "0.0.0.0:4318"  # сюда шлёт auth-svc

exporters:
  otlphttp/tempo:
    endpoint: "http://tempo:4318"  # коллектор шлёт в Tempo

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [otlphttp/tempo]
```

**Шаг 4.** Добавить в `docker-compose.yml` четыре сервиса:
- `otel-collector` — image `otel/opentelemetry-collector-contrib`, монтирует `otelcol-config.yaml`
- `prometheus` — image `prom/prometheus`, монтирует `prometheus.yml`
- `tempo` — image `grafana/tempo`, принимает OTLP от коллектора
- `grafana` — image `grafana/grafana`, монтирует `provisioning/`

**Проверка фазы 1:** `docker compose up otel-collector prometheus grafana tempo` → Grafana открывается на `:3000`, видит оба datasource.

### Фаза 2 — OTEL в Go (инициализация)

**Шаг 4.** Добавить зависимости (`go get`):
```
go.opentelemetry.io/otel
go.opentelemetry.io/otel/sdk/trace
go.opentelemetry.io/otel/sdk/metric
go.opentelemetry.io/otel/exporters/prometheus
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc
github.com/prometheus/client_golang
```

**Шаг 5.** Написать `internal/telemetry/telemetry.go` — функция `Init`:
- Создаёт `otlptracehttp.Exporter` (отправляет spans в **OTEL Collector**, не напрямую в Tempo)
- Создаёт `TracerProvider` с этим экспортером + `BatchSpanProcessor`
- Создаёт `prometheus.Exporter` (метрики в Prometheus формате)
- Создаёт `MeterProvider` с этим экспортером
- Регистрирует оба как глобальные (`otel.SetTracerProvider`, `otelmetric.SetMeterProvider`)
- Возвращает `shutdown func(ctx)` — вызывается при остановке приложения (flush всех буферов)

**Шаг 6.** Добавить в `config.yaml` секцию `otel`:
```yaml
otel:
  collector_endpoint: "http://localhost:4318"  # OTEL Collector, не Tempo напрямую
  sampling_ratio: 1.0
```

**Шаг 7.** Вызвать `telemetry.Init` в `internal/app/run.go` самым первым шагом, зарегистрировать `shutdown` в defer.

**Шаг 8.** Добавить эндпоинт `GET /metrics` на HTTP сервер — `promhttp.Handler()`. Prometheus будет его скрейпить. Этот роут не должен проходить через auth middleware.

**Проверка фазы 2:** `curl localhost:8000/metrics` → видно `go_*` метрики и `process_*` (базовые, Go runtime).

### Фаза 3 — Автоматическая инструментация

**Шаг 9.** HTTP middleware: обернуть chi router в `otelhttp.NewHandler(router, "auth-svc")`.

Что даёт автоматически:
- Span на каждый HTTP запрос (route, method, status code)
- Метрики: `http.server.request.duration` (histogram), `http.server.active_requests` (gauge)
- Propagation: читает `traceparent` заголовок от клиента (связывает трейсы между сервисами)

**Шаг 10.** gRPC interceptors: при создании gRPC сервера добавить:
- `otelgrpc.UnaryServerInterceptor()` — для unary вызовов
- `otelgrpc.StreamServerInterceptor()` — для streaming

**Проверка фазы 3:** сделать любой запрос к API → в Grafana → Tempo найти трейс с spans.

### Фаза 4 — Бизнес-метрики (ручные счётчики)

**Шаг 11.** Добавить кастомные метрики в модули через `otel/metric` API. Метрики создаются один раз при инициализации сервиса, инкрементируются при каждой операции.

Что считаем:

| Метрика | Тип | Labels | Где |
|---------|-----|--------|-----|
| `auth.logins_total` | counter | `method=password\|qr`, `status=ok\|fail` | `session.Service` |
| `auth.registrations_total` | counter | `status=ok\|fail` | `account.Service` |
| `auth.sessions_created_total` | counter | — | `session.Service` |
| `auth.cache_operations_total` | counter | `entity=account\|session\|email`, `result=hit\|miss` | caches |

**Шаг 12.** (опционально) Добавить spans вручную на критичные операции — например, span на `tx.Transaction()` чтобы видеть время транзакции в трейсе.

### Фаза 5 — Grafana дашборд

**Шаг 13.** Создать `observability/grafana/dashboards/auth-svc.json` с панелями:
- RPS (requests per second) по роутам
- Latency p50/p95/p99
- Error rate (5xx)
- Cache hit rate (hits / (hits + misses))
- Активные сессии (registrations_total - deletions_total)

Дашборд провижонируется автоматически при старте Grafana — не нужно создавать руками.

### Порядок выполнения шагов

```
Шаг 1-4  (docker: collector+prometheus+tempo+grafana)
→ Шаг 5 (go get)
→ Шаг 6-9 (telemetry init + /metrics endpoint)
→ Шаг 10-11 (auto instrumentation: http + grpc)
→ Шаг 12 (business metrics)
→ Шаг 13-14 (dashboard)
```

Нумерация шагов выше сдвинулась из-за добавления шага с коллектором — при реализации следить по описанию шагов, не по номерам.

Каждый шаг верифицируется отдельно перед переходом к следующему.

---

## Roadmap

### Следующие шаги (функциональные, в порядке приоритета)

1. **Kafka Consumer + Retry система** (`internal/messenger/`)
   - Consumer читает из Kafka, вызывает хендлеры напрямую (без InboxWorker)
   - Retry через outbox_events + экспоненциальный backoff топики (`{topic}.retry.1m`, `.2m`, `.4m`... `.60m`)
   - DLQ топик при исчерпании попыток
   - evcontroller хендлеры: OrgCreated, OrgDeleted, OrgMemberCreated, OrgMemberDeleted

2. **Тесты**
   - Интеграционные тесты модулей (account, session, auth) — реальная БД, не моки
   - Тесты Consumer + retry логики

### Инфраструктура (не функциональные, после функциональных)

3. **Nginx** — reverse proxy перед REST/gRPC
4. **Kubernetes** — манифесты деплоя (план ниже)

---

## Kubernetes — план деплоя

### Docker Compose → Kubernetes маппинг

```
docker-compose service  →  Deployment + Service
container               →  Pod (минимальная единица запуска)
ports:                  →  Service (ClusterIP / NodePort / LoadBalancer)
environment:            →  ConfigMap + Secret
volumes:                →  PersistentVolumeClaim (PVC)
networks:               →  Namespace + Service DNS
depends_on:             →  readinessProbe / initContainers
```

### Структура файлов

```
k8s/
  namespace.yaml
  auth-svc/
    deployment.yaml
    service.yaml
    configmap.yaml
    secret.yaml
  postgres/
    statefulset.yaml
    service.yaml
    pvc.yaml
  redis/
    statefulset.yaml
    service.yaml
    pvc.yaml
  observability/
    prometheus/
    grafana/
    tempo/
    otel-collector/
  ingress.yaml          ← внешний трафик (nginx-ingress)
```

### Ключевые концепции

**Namespace** — изоляция всех ресурсов проекта:
```bash
kubectl create namespace netbill
kubectl config set-context --current --namespace=netbill
```

**Secret** — sensitive данные (DSN, пароли):
```bash
kubectl create secret generic auth-svc-secrets \
  --from-literal=DATABASE_SQL_URL="postgres://postgres:postgres@postgres:5432/postgres" \
  --from-literal=REDIS_ADDR="redis:6379" \
  --from-literal=REDIS_PASSWORD=""
```

**ConfigMap** — нечувствительные конфиги:
```bash
kubectl create configmap auth-svc-config \
  --from-literal=OTEL_COLLECTOR_ENDPOINT="http://otel-collector:4318"
```

**Service DNS** — внутри кластера сервисы резолвятся как `<name>.<namespace>.svc.cluster.local` или просто `<name>` в том же namespace. Env-переменные в манифестах должны указывать на имена сервисов, не на `localhost`.