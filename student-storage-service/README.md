# Student Storage Service

Это сервис хранения данных телеметрии студентов. Он получает события с помощью Kafka, сохраняет их в PostgreSQL и предоставляет HTTP API для получения данных.

## Структура проекта

```
student-storage-service/
├── internal/
│   ├── config/           # Конфигурация приложения
│   ├── db/               # Работа с базой данных
│   ├── handlers/         # HTTP обработчики
│   └── kafka/            # Обработка сообщений Kafka
├── migrations/           # Миграции базы данных
├── main.go               # Точка входа приложения
├── go.mod                # Зависимости Go
└── go.sum                # Проверка зависимостей
```

## Конфигурация через переменные окружения

| Переменная | Описание | Значение по умолчанию |
|-----------|---------|---------------------|
| `DATABASE_URL` | Для подключения к PostgreSQL | `postgres://postgres:postgres@postgres:5432/studentdata?sslmode=disable` |
| `KAFKA_BROKER` | Адрес Kafka брокера | `kafka:9092` |
| `KAFKA_TOPIC` | Тема Kafka для чтения | `telemetry-events` |
| `KAFKA_GROUP_ID` | Идентификатор группы потребителей Kafka | `student-storage-service` |
| `HTTP_ADDRESS` | Адрес для HTTP сервера | `:8080` |
| `MAX_RETRY_ATTEMPTS` | Максимальное количество попыток подключения к БД | `10` |
| `RETRY_INTERVAL_SECONDS` | Интервал между попытками подключения к БД (секунды) | `3` |

## Миграции базы данных

Проект использует Liquibase для управления миграциями базы данных. Все миграции находятся в директории `migrations/`.

Миграция (`V1__create_tables.sql`) создает:

- Таблицы: `cursor_positions`, `key_presses`, `logs`
- Составные индексы на (session_id, email) для каждой таблицы для оптимизации запросов с фильтрацией



## Endpoints API

- `GET /cursors?session_id=...&email=...&limit=...&offset=...` - получить позиции курсора с фильтрацией и пагинацией
- `GET /keypresses?session_id=...&email=...&limit=...&offset=...` - получить нажатия клавиш с фильтрацией и пагинацией
- `GET /logs?session_id=...&email=...&limit=...&offset=...` - получить логи с фильтрацией и пагинацией
- `GET /session-events?session_id=...` - получить все уникальные email с типом и временем последнего события для указанной сессии

Обязательные параметры:
- `session_id`: идентификатор экзамена
- `email`: email пользователя

Опциональные параметры:
- `limit`: количество записей на странице (по умолчанию 50, максимум 1000)
- `offset`: смещение (по умолчанию 0)
