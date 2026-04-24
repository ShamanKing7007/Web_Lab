# Web_Labs

RESTful API для заметок с аутентификацией и авторизацией на Go (`Gin + GORM + PostgreSQL`).

Проект включает:
- CRUD для заметок
- soft delete
- пагинацию
- регистрацию и вход по email/password
- JWT access/refresh tokens в `HttpOnly` cookies
- refresh-сессии с хранением в БД
- OAuth 2.0 через Yandex ID
- восстановление пароля через одноразовый reset token

## Стек

- Go 1.26
- Gin
- GORM
- PostgreSQL 16
- Docker / Docker Compose
- JWT (`github.com/golang-jwt/jwt/v5`)
- bcrypt

## Запуск

### Через Docker Compose

```bash
cd deploy
cp .env.example .env
docker compose up --build
```

Приложение будет доступно на `http://localhost:4200`.

Дополнительно:
- PostgreSQL: `localhost:5432`
- pgAdmin: `http://localhost:5050`

Остановка:

```bash
cd deploy
docker compose down
```

Полная очистка:

```bash
cd deploy
docker compose down -v
```

## Переменные окружения

Основные переменные из `deploy/.env.example`:

| Переменная | Описание |
|---|---|
| `DB_HOST` | Хост PostgreSQL |
| `DB_PORT` | Порт PostgreSQL |
| `DB_USER` | Пользователь PostgreSQL |
| `DB_PASSWORD` | Пароль PostgreSQL |
| `DB_NAME` | Имя базы данных |
| `PORT` | Порт приложения |
| `JWT_ACCESS_SECRET` | Секрет для access token |
| `JWT_REFRESH_SECRET` | Секрет для refresh token |
| `CLIENT_ID` | Yandex OAuth client id |
| `CLIENT_SECRET` | Yandex OAuth client secret |
| `CALLBACK_URL` | Callback URL OAuth |
| `OAUTH_PROVIDER` | Значение в примере есть, но текущий код поддерживает только `yandex` |
| `PGADMIN_PASSWORD` | Пароль для pgAdmin |

Примечание:
- В `deploy/.env.example` также есть `JWT_ACCESS_EXPIRATION` и `JWT_REFRESH_EXPIRATION`, но текущая версия кода их не читает. Сейчас сроки жизни токенов зафиксированы в коде: `15 минут` и `7 дней`.

## API

### Health

| Метод | URI | Описание |
|---|---|---|
| `GET` | `/health` | Проверка работоспособности |

### Auth

| Метод | URI | Описание | Доступ |
|---|---|---|---|
| `POST` | `/auth/register` | Регистрация пользователя | Public |
| `POST` | `/auth/login` | Вход, установка access/refresh cookies | Public |
| `POST` | `/auth/refresh` | Обновление пары токенов по refresh cookie | Public |
| `GET` | `/auth/whoami` | Получение текущего пользователя | Private |
| `POST` | `/auth/logout` | Завершение текущей сессии | Private |
| `POST` | `/auth/logout-all` | Завершение всех сессий пользователя | Private |
| `GET` | `/auth/oauth/yandex` | Начало OAuth flow | Public |
| `GET` | `/auth/oauth/yandex/callback` | OAuth callback | Public |
| `POST` | `/auth/forgot-password` | Создание reset token | Public |
| `POST` | `/auth/reset-password` | Смена пароля по reset token | Public |

### Notes

Все маршруты `/notes/*` защищены access token из cookie.

| Метод | URI | Описание |
|---|---|---|
| `GET` | `/notes?page=1&limit=10` | Список заметок пользователя |
| `GET` | `/notes/:id` | Получение заметки по ID |
| `POST` | `/notes` | Создание заметки |
| `PUT` | `/notes/:id` | Полное обновление заметки |
| `PATCH` | `/notes/:id` | Частичное обновление заметки |
| `DELETE` | `/notes/:id` | Soft delete заметки |

## Примеры запросов

### Регистрация

```bash
curl -X POST http://localhost:4200/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"user@example.com\",\"password\":\"mypassword\"}"
```

### Вход

```bash
curl -X POST http://localhost:4200/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"user@example.com\",\"password\":\"mypassword\"}" \
  -c cookies.txt
```

### Проверка авторизации

```bash
curl -X GET http://localhost:4200/auth/whoami \
  -b cookies.txt
```

### Список заметок

```bash
curl -X GET "http://localhost:4200/notes?page=1&limit=10" \
  -b cookies.txt
```

### Создание заметки

```bash
curl -X POST http://localhost:4200/notes \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d "{\"title\":\"Первая заметка\",\"content\":\"Текст заметки\"}"
```

### Обновление токенов

```bash
curl -X POST http://localhost:4200/auth/refresh \
  -b cookies.txt \
  -c cookies.txt
```

### Logout текущей сессии

```bash
curl -X POST http://localhost:4200/auth/logout \
  -b cookies.txt \
  -c cookies.txt
```

### Logout всех сессий

```bash
curl -X POST http://localhost:4200/auth/logout-all \
  -b cookies.txt \
  -c cookies.txt
```

### Запрос reset token

Текущая учебная реализация возвращает `reset_token` прямо в ответе API.

```bash
curl -X POST http://localhost:4200/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"user@example.com\"}"
```

### Смена пароля

```bash
curl -X POST http://localhost:4200/auth/reset-password \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"RESET_TOKEN\",\"password\":\"newpassword\"}"
```

## Структура проекта

```text
.
├── deploy/
│   ├── .env.example
│   ├── docker-compose.yml
│   └── Dockerfile
├── internal/
│   ├── apperrors/
│   ├── config/
│   ├── database/
│   ├── notes/
│   │   ├── handler/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── routes/
│   │   ├── service/
│   │   └── validator/
│   └── users/
│       ├── crypto/
│       ├── dto/
│       ├── handler/
│       ├── middleware/
│       ├── models/
│       ├── oauth/
│       ├── repository/
│       ├── routes/
│       └── service/
├── Lab2.md
├── Lab3.md
├── PROTECTION.md
├── main.go
└── README.md
```

## Модели

### User

- `id`
- `email`
- `password_hash`
- `salt`
- `yandex_id`
- `vk_id`
- `reset_token_hash`
- `reset_token_expires_at`
- `created_at`
- `updated_at`
- `deleted_at`

### Token

- `id`
- `user_id`
- `token_hash`
- `expires_at`
- `revoked`
- `created_at`
- `updated_at`
- `deleted_at`

### Note

- `id`
- `user_id`
- `title`
- `content`
- `created_at`
- `updated_at`
- `deleted_at`

## База данных и миграции

При старте приложения выполняется `GORM AutoMigrate` для моделей:
- `Note`
- `User`
- `Token`

Это означает, что для добавления новых nullable-полей обычно достаточно перезапустить приложение.

## Безопасность

- Пароли хешируются через `bcrypt` с отдельной случайной солью.
- Access token хранится в `HttpOnly` cookie.
- Refresh token хранится в `HttpOnly` cookie, а его bcrypt-хеш сохраняется в БД.
- `/auth/logout` отзывает только текущую refresh-сессию.
- `/auth/logout-all` отзывает все refresh-сессии пользователя.
- При `refresh` старый refresh token инвалидируется, создаётся новая пара токенов и новая запись сессии.
- OAuth flow использует параметр `state`.
- Все заметки привязаны к владельцу, проверка прав выполняется в сервисном слое.

Важно:
- В текущей версии cookies выставляются с `HttpOnly`, но без явной настройки `SameSite` и `Secure`.
- Восстановление пароля реализовано как учебный сценарий: reset token возвращается в JSON-ответе, а не отправляется по email.

## Обработка ошибок

Типичные ответы API:
- `400 Bad Request` — некорректные входные данные
- `401 Unauthorized` — отсутствует или невалиден токен
- `403 Forbidden` — попытка доступа к чужому ресурсу
- `404 Not Found` — ресурс не найден
- `409 Conflict` — пользователь с таким email уже существует
- `500 Internal Server Error` — внутренняя ошибка сервера
