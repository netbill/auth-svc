# auth-svc — Claude Context

## Проект

Go микросервис аутентификации (`github.com/netbill/auth-svc`).
Отвечает за регистрацию, логин, сессии, управление аккаунтами.

## Стек

- **Go 1.25**, chi router, pgx v5, squirrel
- **PostgreSQL** — основная БД, миграции через `sql-migrate`
- **Kafka** — события через `github.com/netbill/eventbox` (outbox/inbox pattern)
- **Debezium** — подключён (ветка `feat/debezium`), читает WAL → Kafka через Outbox Event Router SMT

## Структура

```
cmd/auth-svc/          — entrypoint
internal/
  app/                 — запуск приложения (run.go, migrations.go, events.go)
  core/
    auth/              — бизнес-логика аутентификации
    organization/      — бизнес-логика организаций
  messenger/
    consumer.go        — Kafka consumer (читает OrganizationsV1, OrgMembersV1)
    producer.go        — Kafka producer (пишет AccountsTopicV1)
    inbox_worker.go    — InboxWorker (УДАЛЯЕТСЯ, см. ниже)
    outbox_worker.go   — OutboxWorker (УДАЛЯЕТСЯ, см. ниже)
    publisher/         — пишет события в outbox таблицу
    evcontroller/      — хендлеры входящих событий (org, org_member)
  repository/          — интерфейсы репозиториев
  repository/pg/       — PostgreSQL реализации
  rest/                — HTTP сервер, контроллеры, middleware
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

**Открытый вопрос:** retry топик требует backoff (задержку перед повтором). Kafka FIFO не умеет ждать — нужно решить как реализовать задержку (отдельный топик на каждый уровень retry, или timestamp в сообщении + проверка в Consumer).

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

## Зависимости eventbox

- `github.com/netbill/eventbox v0.1.14` — outbox/inbox interfaces + workers
- `github.com/netbill/evtypes v0.1.3` — топики и типы событий
- `github.com/netbill/pgdbx v0.3.1` — PostgreSQL helpers

## Конфигурация

`config.yaml` / `config.example.yaml` — основной конфиг через viper.
`config_docker.yaml` — для docker окружения.

## Команды

```bash
# запуск
go run ./cmd/auth-svc

# миграции
go run ./cmd/auth-svc migrate-up
go run ./cmd/auth-svc migrate-down
```
