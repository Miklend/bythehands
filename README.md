# Taalkbout / TalkaBot (MVP scaffold)

Репозиторий содержит **два отдельных микросервиса**:

- `taalkbout-api` — core REST API + бизнес-логика + PostgreSQL.
- `taalkbout-bot` — Telegram bot service без бизнес-логики, общается с API **только по HTTP**.

## Быстрый старт

1) Создать `.env`:

```bash
cp .env.example .env
```

2) Указать реальный `TELEGRAM_BOT_TOKEN` в `.env`.

3) Запуск:

```bash
docker compose up --build
```

При старте `taalkbout-api` автоматически применяет миграции из `./migrations`.

Логи:

```bash
docker compose logs -f taalkbout-api
docker compose logs -f taalkbout-bot
```

Остановка:

```bash
docker compose down
```

## Миграции (golang-migrate внутри `taalkbout-api`)

Поднять миграции:

```bash
docker compose exec taalkbout-api migrate -path /migrations -database "$DATABASE_URL" up
```

Откатить миграции:

```bash
docker compose exec taalkbout-api migrate -path /migrations -database "$DATABASE_URL" down
```

## API

Base URL (из compose): `http://localhost:8080`

Health:

```bash
curl -s http://localhost:8080/health | jq
```

### Users

Создать/найти пользователя по `telegram_id`:

```bash
curl -s -X POST http://localhost:8080/api/v1/users/telegram \
  -H 'Content-Type: application/json' \
  -d '{"telegram_id":123456789,"username":"alice","display_name":"Alice"}' | jq
```

### Pairs

Создать пару (в ответе будет `invite_token`):

```bash
curl -s -X POST http://localhost:8080/api/v1/pairs \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"<USER_UUID>"}' | jq
```

Получить пару:

```bash
curl -s http://localhost:8080/api/v1/pairs/<PAIR_UUID> | jq
```

Присоединиться по invite token:

```bash
curl -s -X POST http://localhost:8080/api/v1/invites/<INVITE_TOKEN>/join \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"<USER_UUID>"}' | jq
```

### Issues

Создать тему:

```bash
curl -s -X POST http://localhost:8080/api/v1/pairs/<PAIR_UUID>/issues \
  -H 'Content-Type: application/json' \
  -d '{
    "created_by_user_id":"<USER_UUID>",
    "title":"Тема",
    "description":"Описание",
    "priority":"medium",
    "visibility":"visible",
    "repeat_threshold":2
  }' | jq
```

Список тем:

```bash
curl -s "http://localhost:8080/api/v1/pairs/<PAIR_UUID>/issues?status=active" | jq
```

Карточка темы:

```bash
curl -s http://localhost:8080/api/v1/issues/<ISSUE_UUID> | jq
```

Отметить повторение:

```bash
curl -s -X POST http://localhost:8080/api/v1/issues/<ISSUE_UUID>/repeat \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"<USER_UUID>","note":"опять"}' | jq
```

Изменить статус:

```bash
curl -s -X PATCH http://localhost:8080/api/v1/issues/<ISSUE_UUID>/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"resolved"}' | jq
```

### Conversations

Начать разговор:

```bash
curl -s -X POST http://localhost:8080/api/v1/issues/<ISSUE_UUID>/conversations \
  -H 'Content-Type: application/json' \
  -d '{"pair_id":"<PAIR_UUID>","goal":"обсудить спокойно"}' | jq
```

Завершить:

```bash
curl -s -X PATCH http://localhost:8080/api/v1/conversations/<CONV_UUID>/finish \
  -H 'Content-Type: application/json' \
  -d '{"result_status":"resolved","result_text":"договорились"}' | jq
```

Пауза/продолжить:

```bash
curl -s -X PATCH http://localhost:8080/api/v1/conversations/<CONV_UUID>/pause | jq
curl -s -X PATCH http://localhost:8080/api/v1/conversations/<CONV_UUID>/resume | jq
```

История по паре:

```bash
curl -s "http://localhost:8080/api/v1/pairs/<PAIR_UUID>/conversations?status=finished" | jq
```

## Test mode (API)

Включить:

```bash
curl -s -X POST http://localhost:8080/api/v1/test-mode/start \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"<USER_UUID>"}' | jq
```

Выключить:

```bash
curl -s -X POST http://localhost:8080/api/v1/test-mode/stop \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"<USER_UUID>"}' | jq
```

## Bot

Команды бота (MVP):

- `/start` — onboarding + deep-link join по invite token
- `/menu` — главное меню
- `/topics` — список активных тем
- `/focus` — подготовка и запуск разговора по теме
- `/history` — история завершенных разговоров
- `/settings` — правила + выбор пространства
- `/testmode` — включить/выключить тестовый режим

Тестовый режим:
- создаёт отдельную тестовую пару с виртуальным партнером
- не влияет на “реальную” пару
- позволяет одному человеку прогнать весь UX
