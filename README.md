# Dating Backend

Backend API для мобильного приложения знакомств Dating.

## Что внутри

- HTTP API на `echo` (`/api/v1/...`)
- PostgreSQL/PostGIS
- Redis
- S3-совместимое хранилище (MinIO)
- JWT-авторизация, WebSocket-чат
- Метрики Prometheus (`/api/v1/metrics`)
- Интеграции: YooKassa, VK ID, Firebase (опционально)

## Требования

- Go `1.21+`
- Docker + Docker Compose
- `make`
- Опционально: `golang-migrate`, `golangci-lint`

## Быстрый старт (рекомендуется для разработки)

1. Создать env-файл:

```bash
cp .example.env .env
```

2. Проверить и заполнить обязательные переменные в `.env`:

- `VK_APP_ID` (в `.example.env` отсутствует, но обязателен)
- `YOUKASSA_SECRET_KEY`
- при необходимости `FIREBASE_*`, `SMTP_*`

3. Поднять инфраструктуру локально:

```bash
docker compose -f compose.dev.yml up -d
```

4. Применить миграции:

```bash
make migrate-up PG_URL=postgres://postgres:mysecretpassword@localhost:5432/postgres
```

Если `golang-migrate` не установлен, можно вместо шагов 4-5 выполнить:

```bash
go run -tags migrate ./cmd/app
```

5. Запустить приложение:

```bash
go run ./cmd/app
```

API будет доступно по адресу: `http://localhost:8032/api/v1`

## Запуск полностью в Docker

Для запуска сервиса `app` в контейнере значения в `.env` должны быть сетевыми именами сервисов из compose:

- `PG_HOST=postgis`
- `RDB_ADDR=redis:6379`
- `S3_ENDPOINT=minio:9000`

После этого:

```bash
docker compose up --build -d
```

## Полезные команды

```bash
make help
make build
make lint
make migrate-create NAME=add_new_table
make migrate-up PG_URL=postgres://postgres:mysecretpassword@localhost:5432/postgres
make migrate-down PG_URL=postgres://postgres:mysecretpassword@localhost:5432/postgres
```

Создание админа:

```bash
go run ./cmd/cli create-admin <login> <password>
```

Интеграционные тесты:

```bash
docker compose -f integration_tests/compose.yml up --build --abort-on-container-exit --exit-code-from app-integration
```

## API и документация

- OpenAPI-спека: `doc/swagger.yml`
- Базовый URL локально: `http://localhost:8032/api/v1`
- Healthcheck: `GET /api/v1/health`
- Метрики: `GET /api/v1/metrics`

## Структура проекта

- `cmd/app` - основной HTTP-сервис
- `cmd/cli` - служебная CLI-утилита (например, создание администратора)
- `internal/service` - бизнес-логика
- `internal/transport/http` - HTTP-слой и роутинг
- `internal/service/repo` - доступ к данным
- `migrations` - SQL-миграции
- `integration_tests` - интеграционные тесты в отдельном compose-контуре
