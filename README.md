# WoodenButWithSoul API — Магазин изделий из дерева

RESTful API для управления пользователями магазина изделий из дерева.

## Требования

- Go 1.26
- Docker и Docker Compose
- PostgreSQL 16+

## Старт

### Запуск через Docker Compose

```bash
# Сборка и запуск контейнеров
docker-compose up --build

# Запуск в фоновом режиме
docker-compose up -d --build

# Просмотр логов
docker-compose logs -f app

# Остановка контейнеров
docker-compose down

# Полная очистка (контейнеры + volumes)
docker-compose down -v
```

### Запуск без Docker

```bash
# Загрузка зависимостей
go mod download

# Сборка приложения
go build -o main .

# Запуск сервера
go run main.go
```

## API Endpoints

| Метод | URI | Описание | Статус |
|-------|-----|----------|--------|
| `GET` | `/info` | Дни до Нового года | `200 OK` |
| `GET` | `/users` | Список пользователей (с пагинацией) | `200 OK` |
| `GET` | `/users/:id` | Пользователь по ID | `200 OK` / `404 Not Found` |
| `POST` | `/users` | Создать пользователя | `201 Created` / `409 Conflict` |
| `PUT` | `/users/:id` | Полное обновление | `200 OK` / `404 Not Found` |
| `PATCH` | `/users/:id` | Частичное обновление | `200 OK` / `404 Not Found` |
| `DELETE` | `/users/:id` | Удалить (Soft Delete) | `204 No Content` / `404 Not Found` |


## Структура проекта

```
.
├── internal/
│   ├── config/          # Конфигурация приложения
│   ├── controllers/     # HTTP контроллеры
│   ├── errors/          # Кастомные ошибки
│   ├── handler/         # Роутер и обработчики
│   ├── models/          # Модели данных
│   ├── repository/      # Репозиторий (работа с БД)
│   ├── service/         # Бизнес-логика
│   └── validator/       # Валидация данных
├── .env                 # Переменные окружения
├── .env.example         # Шаблон переменных
├── docker-compose.yml   # Docker Compose конфигурация
├── Dockerfile           # Docker образ
├── go.mod               # Go модуль
├── go.sum               # Go зависимости
├── main.go              # Точка входа
└── README.md            # Документация
```

## База данных

### Модель User

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ |
| `name` | VARCHAR(100) | Имя пользователя |
| `email` | VARCHAR(255) | Email (уникальный) |
| `password` | VARCHAR(255) | Хешированный пароль (bcrypt) |
| `created_at` | TIMESTAMP | Дата создания |
| `updated_at` | TIMESTAMP | Дата обновления |
| `deleted_at` | TIMESTAMP | Soft delete (NULL = активен) |

### Миграции

Приложение использует **GORM AutoMigrate** для автоматического создания и обновления схемы базы данных при запуске.

## Безопасность и валидация

- Пароли хешируются с помощью **bcrypt**
- Soft delete для сохранения истории данных
- Валидация входных данных:
  - Email: формат email
  - Пароль: минимум 4 символа
  - Имя: 2-100 символов
  - Пагинация: page >= 1, limit 1-100

## Зависимости

- [Gin](https://github.com/gin-gonic/gin) — HTTP фреймворк
- [GORM](https://gorm.io/) — ORM для Go
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) — Хеширование паролей
- [UUID](https://github.com/google/uuid) — Генерация UUID
