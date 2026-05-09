# Student Data Service

Это сервис приёма данных студентов. Он принимает события по HTTP и публикует их в Kafka для дальнейшей обработки.

## Структура проекта

```
student-data-service/
├── internal/
│   ├── config/           # Конфигурация приложения
│   ├── handlers/         # HTTP обработчики
│   └── kafka/            # Kafka продюсер
├── main.go               # Точка входа приложения
├── go.mod                # Зависимости Go
└── go.sum                # Проверка зависимостей
```

## Конфигурация через переменные окружения

| Переменная | Описание | Значение по умолчанию |
|-----------|---------|---------------------|
| `KAFKA_BROKER` | Адрес Kafka брокера | `kafka:9092` |
| `KAFKA_TOPIC` | Топик Kafka для публикации | `telemetry-events` |
| `HTTP_ADDRESS` | Адрес для HTTP сервера | `:5000` |

## Endpoints API

- `POST /telemetry` — принять событие телеметрии и опубликовать его в Kafka

### Формат запроса

```json
{
  "event": "cursor_position",
  "timestamp": "2024-01-01T00:00:00Z",
  "sessionId": "session-123",
  "customUserId": "user@example.com",
  "data": { ... }
}
```

Обязательные поля: `event`, `timestamp`, `sessionId`, `customUserId`, `data`.

### Поддерживаемые типы событий

| Тип события | Поля `data` |
|------------|------------|
| `cursor_position` | `x`, `y`, `ts` (обязательное) |
| `key_press` | `keyCode`, `keyName` (обязательное), `modifiers`, `isCombo`, `ts` (обязательное) |
| `log` | `level` (обязательное), `message` (обязательное), `ts` (обязательное) |
