# Web_Labs

RESTful API для заметок с аутентификацией и авторизацией на Go (`Gin + GORM + PostgreSQL`).

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

Swagger UI в режиме разработки доступен на `http://localhost:4200/api/docs/index.html`.
Если `APP_ENV=production` или `SWAGGER_ENABLED=false`, маршрут документации не регистрируется и возвращает `404 Not Found`.

Дополнительно:
- PostgreSQL: `localhost:5432`
- pgAdmin: `http://localhost:5050`

Остановка:

```bash
cd deploy
docker compose down
```
## Переменные окружения

Основные переменные из `deploy/.env.example`:

| Переменная | Описание |
|---|---|
| `DB_HOST` | Хост PostgreSQL, по умолчанию `localhost` |
| `DB_PORT` | Порт PostgreSQL, по умолчанию `5432` |
| `DB_USER` | Пользователь PostgreSQL, обязательная переменная |
| `DB_PASSWORD` | Пароль PostgreSQL, обязательная переменная |
| `DB_NAME` | Имя базы данных, по умолчанию `Web_Labs` |
| `PORT` | Порт приложения, по умолчанию `4200` |
| `APP_ENV` | Окружение приложения; влияет на Swagger (`production` отключает его по умолчанию) |
| `SWAGGER_ENABLED` | Явное включение или выключение Swagger UI |
| `JWT_ACCESS_SECRET` | Секрет для access token, обязательная переменная |
| `JWT_REFRESH_SECRET` | Секрет для refresh token, обязательная переменная |
| `JWT_ACCESS_EXPIRATION` | Время жизни access token JWT, например `30s`, `15m`, `1h` |
| `JWT_REFRESH_EXPIRATION` | Время жизни refresh token JWT; поддерживается и формат дней, например `7d` |
| `CLIENT_ID` | Yandex OAuth client id |
| `CLIENT_SECRET` | Yandex OAuth client secret |
| `CALLBACK_URL` | Callback URL OAuth, например `http://localhost:4200/auth/oauth/yandex/callback` |
| `OAUTH_PROVIDER` | Значение есть в конфиге, но текущая реализация обрабатывает только `yandex` |
| `PGADMIN_PASSWORD` | Пароль для pgAdmin |

Примечание:
- `JWT_ACCESS_EXPIRATION` и `JWT_REFRESH_EXPIRATION` читаются приложением из `.env`.
- Поддерживаются стандартные значения Go duration: `30s`, `15m`, `1h`.
- Для refresh token дополнительно поддержан формат с днями, например `7d`.
- Время жизни cookies `access_token` и `refresh_token` синхронизировано с `JWT_ACCESS_EXPIRATION` и `JWT_REFRESH_EXPIRATION`.
- Если `APP_ENV` не задан, приложение использует `development`. Также поддерживается fallback на `NODE_ENV`.
- Для локальной разработки используйте `APP_ENV=development` и `SWAGGER_ENABLED=true`.


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
| `DELETE` | `/notes/:id` | Soft delete заметки, ответ `204 No Content` |

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

### OAuth через Yandex

```bash
curl -i http://localhost:4200/auth/oauth/yandex
```

## Структура проекта

```text
.
├── deploy/
│   ├── .env.example
│   ├── docker-compose.yml
│   └── Dockerfile
├── docs/
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── apperrors/
│   ├── config/
│   ├── database/
│   ├── httpapi/
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
- `type`
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

## Обработка ошибок

Типичные ответы API:
- `400 Bad Request` — некорректные входные данные
- `401 Unauthorized` — отсутствует или невалиден токен
- `403 Forbidden` — попытка доступа к чужому ресурсу
- `404 Not Found` — ресурс не найден
- `409 Conflict` — пользователь с таким email уже существует
- `500 Internal Server Error` — внутренняя ошибка сервера

## Ограничения текущей реализации

- Основной runtime использует HttpOnly cookies для `access_token` и `refresh_token`; Bearer-схема в Swagger нужна в первую очередь для ручного тестирования.
- OAuth-маршруты параметризованы как `/auth/oauth/:provider`, но обработчик сейчас принимает только `yandex`.
- Поле `vk_id` и связанные структуры в моделях уже есть, но полноценный VK OAuth flow не реализован.
