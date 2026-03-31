# Shortener

Сервис сокращения ссылок на Go. Поддерживает кастомные коды, аналитику визитов и кэширование через Redis.

## Стек

- **Go 1.26** — язык
- **PostgreSQL 16** — хранилище ссылок и визитов
- **Redis 7** — кэш ссылок
- **golang-migrate** — миграции БД
- **pgx/v5** — драйвер PostgreSQL

## Get Started

### Вариант 1 — Docker Compose (рекомендуется)

1. Скопируйте конфиг и при необходимости отредактируйте:

```bash
cp config.example.yaml config.yaml
```

2. Поднимает PostgreSQL, Redis, запускает миграции и само приложение:

```bash
docker compose up -d
```

Приложение будет доступно на `http://localhost:8080`.

Остановить и удалить контейнеры:

```bash
docker compose down
```

---

### Вариант 2 — локальный запуск

**Требования:** Go 1.22+, PostgreSQL, Redis, [Task](https://taskfile.dev).

1. Скопируйте конфиг и при необходимости отредактируйте:

```bash
cp config.example.yaml config.yaml
```

2. Запустите зависимости (PostgreSQL и Redis):

```bash
docker compose up -d postgres redis
```

3. Примените миграции:

```bash
go run ./cmd/app/main.go --config config.yaml --migrate
```

4. Запустите приложение:

```bash
task run
```

---

## API

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/shorten` | Создать короткую ссылку |
| `GET` | `/s/{code}` | Редирект по короткому коду |
| `GET` | `/analytics/{code}` | Аналитика ссылки |
| `GET` | `/` | Веб-интерфейс |

### POST /shorten

```json
{
  "long_url": "https://example.com/very/long/path",
  "custom_code": "mycode"
}
```

`custom_code` — опционально, 3–30 символов, только буквы, цифры и дефис.

### GET /analytics/{code}

```json
{
  "total_visits": 42,
  "by_day": [{ "day": "2026-03-31", "count": 10 }],
  "by_month": [{ "month": "2026-03", "count": 42 }],
  "by_user_agent": [{ "user_agent": "Chrome/120", "count": 30 }]
}
```

---

## Тесты

```bash
# Unit-тесты (без Docker)
go test ./internal/service/... ./internal/handler/...

# Интеграционные тесты (поднимают контейнеры через testcontainers-go)
go test ./internal/repository/... ./internal/cache/redis/...

# Все тесты без кэша
go test -count=1 ./...

# С покрытием
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

---

## Taskfile

| Команда | Описание |
|---|---|
| `task run` | Запустить приложение |
| `task lint` | Запустить golangci-lint |
| `task docker:up` | Поднять все контейнеры |
| `task docker:down` | Остановить контейнеры |
| `task docker:logs` | Логи приложения |
| `task docker:rebuild` | Пересобрать и перезапустить app |
