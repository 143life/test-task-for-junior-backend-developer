Понял, давай дополним существующий README в том же лаконичном стиле, добавив только информацию о новой функциональности. Вот обновлённый файл:

```markdown
# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файлы из `migrations/` монтируются в `docker-entrypoint-initdb.d` и применяются только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

## Периодические задачи

При создании или обновлении задачи можно указать объект `schedule` с правилом повторения. Поддерживаются четыре типа:

| Тип       | Поле        | Описание                                       |
|-----------|-------------|------------------------------------------------|
| `daily`   | `interval`  | Каждые N дней (N > 0)                          |
| `monthly` | `day`       | В определённое число месяца (1–31)             |
| `dates`   | `dates`     | Список конкретных дат в формате `YYYY-MM-DD`   |
| `parity`  | `parity`    | `"even"` — по чётным дням, `"odd"` — по нечётным |

Пример создания ежедневной задачи:

```json
POST /api/v1/tasks
{
  "title": "Обзвон пациентов",
  "status": "new",
  "schedule": {
    "type": "daily",
    "interval": 1
  }
}
```

При наступлении очередной даты выполнения фоновый планировщик автоматически создаёт дочернюю задачу без расписания (одноразовый экземпляр). Планировщик использует механизм `LISTEN`/`NOTIFY` в PostgreSQL для мгновенной реакции на изменения.

## Тестирование периодичности

Для немедленной проверки работы планировщика можно создать задачу с типом `dates` и указать сегодняшнюю дату — дочерняя задача появится сразу. Логи планировщика выводятся в консоль контейнера `app`.

## Структура проекта

- `cmd/api` – точка входа
- `internal/domain/task` – сущности и бизнес-логика расчёта дат
- `internal/usecase/task` – сценарии и валидация
- `internal/repository/postgres` – слой доступа к данным
- `internal/transport/http` – обработчики, роутер, DTO, Swagger
- `internal/scheduler` – фоновый планировщик
- `migrations` – SQL-миграции