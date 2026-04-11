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

## Структура

```
cmd/auth-svc/          — entrypoint
internal/
  app/                 — запуск приложения (run.go, migrations.go, events.go)
  modules/
    auth/              — ValidateSession (переиспользуется в account и session)
    account/           — регистрация, управление аккаунтом
    session/           — логин, сессии, токены; options.go — ListSessionsOption types
  messenger/
    consumer.go        — Kafka consumer (читает OrganizationsV1, OrgMembersV1)
    producer.go        — Kafka producer (пишет AccountsTopicV1)
    publisher/         — пишет события в outbox таблицу
    evcontroller/      — хендлеры входящих событий (org, org_member)
  repo/
    pg/                — PostgreSQL реализации (accounts, emails, passwords, sessions)
    chache/            — Redis реализации (accounts, emails, passwords, sessions)
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

### TTL стратегия — обсуждается, решение не принято

**Зафиксированные наблюдения (не финальные решения):**

- **Lazy TTL опасен для популярных сущностей** — горячий аккаунт/сессия читается постоянно → TTL бесконечно продлевается → данные могут быть неактуальны часами. Чем популярнее сущность, тем хуже работает lazy TTL.
- **Фиксированный TTL** даёт предсказуемое окно неактуальности независимо от нагрузки — по сути SLA на staleness.
- **Текущий код** уже не продлевает TTL при cache hit (Set вызывается только после DB read или DB write). Это соответствует фиксированному TTL если реализация Redis не делает KEEPTTL.

**Проблема инвалидации сессий (TOCTOU race):**
- `DeleteMySessions` удаляет из DB, но горутина параллельного запроса может записать сессию обратно в кеш уже после удаления
- Версионирование через `account.Version` не решает проблему — для проверки версии всё равно нужен DB/cache hit, возвращаемся к исходной проблеме
- Распределённые транзакции PG↔Redis невозможны в принципе, при шардировании (10 PG + 10 Redis) тем более
- **Вывод:** идеальной консистентности между Redis и Postgres не будет. Вопрос в приемлемом размере окна.

**Открытые вопросы по TTL:**
- Какой TTL для сессий? Для аккаунта?
- Принимаем ли eventual consistency для сессионного кеша или меняем стратегию?

### Модули и кеш
- `auth.Service` — читает accountCache + sessionsCache для ValidateSession
- `account.Service` — читает/пишет accountCache, emailCache; не знает про sessionsCache
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
- **scan хелперы** — централизованная обработка ошибок через `switch { case deleted_at != nil → domain error; case ErrNoRows → not found; case err != nil → wrap }`
- **Unique constraint** — `pgconn.PgError` с кодом `23505` → доменная ошибка (`ErrorUsernameAlreadyTaken`, `ErrorEmailAlreadyExist`)
- **Без query builders** — только чистый SQL с плейсхолдерами `$1, $2, ...`

### cache репозитории
- **Ошибка Redis** — `switch { case errors.Is(err, redis.Nil): → ErrCacheMiss; case err != nil: → err }`
- **JSON сериализация** — `json.Marshal/Unmarshal` для всех сущностей

### Option types для list запросов
- Типы `ListSessionsOption`, `ListSessionsOptions`, `WithDeleted`, `WithLimit` и т.д. — **в domain пакете** (`internal/modules/session/options.go`), не в pg
- pg пакет импортирует domain пакет для типов и строит SQL условия инлайн

## Зависимости eventbox

- `github.com/netbill/eventbox v0.1.14` — outbox/inbox interfaces + workers
- `github.com/netbill/evtypes v0.1.3` — топики и типы событий
- `github.com/netbill/pgdbx v0.3.1` — PostgreSQL helpers
- `github.com/redis/go-redis/v9` — Redis клиент (требует `go mod vendor`)

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
