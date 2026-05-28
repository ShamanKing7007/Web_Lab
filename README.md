# Web_Labs

REST API для заметок с аутентификацией, OAuth, Redis-кешем и хранением данных в MongoDB.

## Стек

- Go 1.26
- Gin
- MongoDB 6
- Redis 7
- MinIO
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
- MinIO API: `localhost:9000`
- MinIO Console: `localhost:9001`
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
| `MINIO_ENDPOINT` | Адрес MinIO, в Docker Compose используется `minio:9000` |
| `MINIO_ACCESS_KEY` | Access key для MinIO |
| `MINIO_SECRET_KEY` | Secret key для MinIO |
| `MINIO_BUCKET` | Bucket для хранения файлов |
| `MINIO_USE_SSL` | Использовать SSL для подключения к MinIO |
| `MAX_FILE_SIZE` | Максимальный размер загружаемого файла в байтах, по умолчанию `10485760` |

## MongoDB

PostgreSQL и pgAdmin в Lab6 не используются. Приложение подключается к MongoDB через `MONGO_URI`, а при старте создает индексы для коллекций:

- `users`: уникальный `email`, sparse unique `yandex_id`, sparse unique `vk_id`;
- `tokens`: уникальный `token_hash`, индекс для поиска активных refresh-сессий;
- `notes`: индекс по `user_id`, `deleted_at`, `created_at`.
- `files`: индекс по `user_id`, `deleted_at`, `created_at`, уникальный `object_key`.

Коллекции:

- `users`
- `tokens`
- `notes`
- `files`

Файлы не сохраняются в MongoDB как BLOB. В коллекции `files` хранятся только метаданные: UUID, владелец, оригинальное имя, размер, MIME-type, bucket, object key, даты создания/обновления и `deleted_at`.

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
- `GET /files/:id` использует кеш метаданных файла.
- `DELETE /files/:id` инвалидирует кеш метаданных файла.

Ключи Redis имеют префикс `wp:` и TTL:

```text
wp:notes:user:{userId}:detail:{noteId}
wp:notes:user:{userId}:list:page:{page}:limit:{limit}
wp:users:profile:{userId}
wp:auth:user:{userId}:access:{jti}
wp:files:{fileId}:meta
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

### Files

Все маршруты `/files/*` защищены access token из cookie или заголовка `Authorization: Bearer`.

| Метод | URI | Описание | Доступ |
|---|---|---|---|
| `POST` | `/files` | Потоковая загрузка PNG/JPEG/JPG через `multipart/form-data`, поле `file` | Private |
| `GET` | `/files/:id` | Скачивание файла из MinIO с заголовками `Content-Type`, `Content-Disposition`, `Content-Length` | Только владелец |
| `DELETE` | `/files/:id` | Soft delete метаданных и удаление объекта из MinIO | Только владелец |

Ограничения загрузки:

- разрешенные MIME-type: `image/png`, `image/jpeg`, `image/jpg`;
- максимальный размер: `MAX_FILE_SIZE`, по умолчанию 10 MB;
- файл передается в MinIO потоком, без полного чтения в память приложения.

### Profile

Все маршруты `/profile` защищены access token из cookie или заголовка `Authorization: Bearer`.

| Метод | URI | Описание |
|---|---|---|
| `GET` | `/profile` | Получение текущего профиля |
| `POST` | `/profile` | Обновление `display_name`, `bio`, `avatar_file_id` |

При установке `avatar_file_id` приложение проверяет, что файл принадлежит текущему пользователю. Предыдущий avatar-файл помечается удаленным.

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

Загрузка файла:

```bash
curl -X POST http://localhost:4200/files \
  -b cookies.txt \
  -F "file=@avatar.png"
```

Скачивание файла:

```bash
curl -L http://localhost:4200/files/550e8400-e29b-41d4-a716-446655440000 \
  -b cookies.txt \
  -o avatar.png
```

Обновление профиля с аватаром:

```bash
curl -X POST http://localhost:4200/profile \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d "{\"display_name\":\"Ivan Petrov\",\"bio\":\"Backend developer\",\"avatar_file_id\":\"550e8400-e29b-41d4-a716-446655440000\"}"
```

Получение профиля:

```bash
curl -X GET http://localhost:4200/profile \
  -b cookies.txt
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
│   ├── storage/
│   │   ├── handler/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── routes/
│   │   └── service/
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
