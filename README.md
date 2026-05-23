# Web_Labs

REST API для заметок с аутентификацией, OAuth, Redis-кешем и хранением данных в MongoDB.

## Стек

- Go 1.26
- Gin
- MongoDB 6
- Redis 7
- Docker / Docker Compose
- JWT (`github.com/golang-jwt/jwt/v5`)
- bcrypt
- Swagger UI

## Запуск

```bash
cd deploy
cp .env.example .env
docker compose up --build
```

Приложение будет доступно на `http://localhost:4200`.

Swagger UI в режиме разработки доступен на `http://localhost:4200/api/docs/index.html`. Если `APP_ENV=production` или `SWAGGER_ENABLED=false`, маршрут Swagger не регистрируется.

Сервисы:

- MongoDB: `localhost:27017`
- Redis: `localhost:6379`
- API: `localhost:4200`

Остановка:

```bash
cd deploy
docker compose down
```

## Переменные окружения

Основные переменные находятся в `deploy/.env.example`.

| Переменная | Описание |
|---|---|
| `DB_USER` | Пользователь MongoDB root |
| `DB_PASSWORD` | Пароль MongoDB root |
| `DB_NAME` | Имя базы MongoDB |
| `MONGO_URI` | Строка подключения к MongoDB |
| `REDIS_HOST` | Хост Redis, в Docker Compose используется `redis` |
| `REDIS_PORT` | Порт Redis |
| `REDIS_PASSWORD` | Пароль Redis |
| `REDIS_DB` | Номер базы Redis |
| `CACHE_TTL_DEFAULT` | TTL кеша в секундах или Go duration |
| `PORT` | Порт приложения |
| `APP_ENV` | Окружение приложения |
| `SWAGGER_ENABLED` | Явное включение/выключение Swagger UI |
| `JWT_ACCESS_SECRET` | Секрет access token |
| `JWT_REFRESH_SECRET` | Секрет refresh token |
| `JWT_ACCESS_EXPIRATION` | Время жизни access token, например `15m` |
| `JWT_REFRESH_EXPIRATION` | Время жизни refresh token, например `7d` |
| `CLIENT_ID` | Yandex OAuth client id |
| `CLIENT_SECRET` | Yandex OAuth client secret |
| `CALLBACK_URL` | Callback URL OAuth, например `http://localhost:4200/auth/oauth/yandex/callback` |
| `OAUTH_PROVIDER` | Сейчас обработчик поддерживает `yandex` |

## MongoDB

PostgreSQL и pgAdmin в Lab6 не используются. Приложение подключается к MongoDB через `MONGO_URI`, а при старте создает индексы для коллекций:

- `users`: уникальный `email`, sparse unique `yandex_id`, sparse unique `vk_id`;
- `tokens`: уникальный `token_hash`, индекс для поиска активных refresh-сессий;
- `notes`: индекс по `user_id`, `deleted_at`, `created_at`.

Коллекции:

- `users`
- `tokens`
- `notes`

В API сохраняются текущие UUID-идентификаторы. Soft delete реализован через поле `deleted_at`; запросы чтения фильтруют удаленные документы.

Проверка MongoDB через CLI:

```bash
docker compose exec mongo mongosh -u "$DB_USER" -p "$DB_PASSWORD" --authenticationDatabase admin
use Web_Labs
db.notes.find()
db.users.find()
db.tokens.find()
```

## Redis Cache

Redis используется для cache-aside кеширования и управления access-сессиями:

- `GET /notes/:id` кеширует конкретную заметку текущего пользователя.
- `GET /notes?page=1&limit=10` кеширует список заметок текущего пользователя.
- `GET /auth/whoami` кеширует публичный профиль пользователя.
- `POST /auth/logout` удаляет JTI текущего access token и кеш профиля.
- `POST /auth/logout-all` удаляет все access-JTI пользователя и кеш профиля.
- `POST/PUT/PATCH/DELETE /notes` инвалидируют кеш списков.
- `PUT/PATCH/DELETE /notes/:id` инвалидируют кеш конкретной заметки.

Ключи Redis имеют префикс `wp:` и TTL:

```text
wp:notes:user:{userId}:detail:{noteId}
wp:notes:user:{userId}:list:page:{page}:limit:{limit}
wp:users:profile:{userId}
wp:auth:user:{userId}:access:{jti}
```

Проверка Redis:

```bash
docker compose exec redis redis-cli --pass "$REDIS_PASSWORD" KEYS 'wp:*'
docker compose exec redis redis-cli --pass "$REDIS_PASSWORD" TTL 'wp:users:profile:{userId}'
docker compose exec redis redis-cli --pass "$REDIS_PASSWORD" GET 'wp:auth:user:{userId}:access:{jti}'
```

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

Регистрация:

```bash
curl -X POST http://localhost:4200/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"user@example.com\",\"password\":\"mypassword\"}"
```

Вход:

```bash
curl -X POST http://localhost:4200/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"user@example.com\",\"password\":\"mypassword\"}" \
  -c cookies.txt
```

Проверка авторизации:

```bash
curl -X GET http://localhost:4200/auth/whoami \
  -b cookies.txt
```

Создание заметки:

```bash
curl -X POST http://localhost:4200/notes \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d "{\"title\":\"Первая заметка\",\"content\":\"Текст заметки\"}"
```

Список заметок:

```bash
curl -X GET "http://localhost:4200/notes?page=1&limit=10" \
  -b cookies.txt
```

Logout:

```bash
curl -X POST http://localhost:4200/auth/logout \
  -b cookies.txt \
  -c cookies.txt
```

OAuth через Yandex:

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
│   ├── cache/
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
├── main.go
└── README.md
```

## Ошибки API

- `400 Bad Request` - некорректные входные данные
- `401 Unauthorized` - отсутствует или невалиден токен
- `403 Forbidden` - попытка доступа к чужому ресурсу
- `404 Not Found` - ресурс не найден
- `409 Conflict` - пользователь с таким email уже существует
- `500 Internal Server Error` - внутренняя ошибка сервера

## Ограничения

- Основной runtime использует HttpOnly cookies для `access_token` и `refresh_token`.
- OAuth-маршруты параметризованы как `/auth/oauth/:provider`, но обработчик сейчас принимает только `yandex`.
- Поле `vk_id` есть в модели пользователя, но полноценный VK OAuth flow не реализован.
