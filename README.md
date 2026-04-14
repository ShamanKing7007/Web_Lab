# Web_Labs — RESTful API для заметок

RESTful API для управления заметками, реализованное на Go (Gin + GORM + PostgreSQL).

## Стек технологий

- **Язык:** Go 1.26
- **Фреймворк:** Gin (HTTP-маршрутизация)
- **ORM:** GORM (работа с БД)
- **СУБД:** PostgreSQL 16
- **Миграции:** GORM AutoMigrate
- **Контейнеризация:** Docker + Docker Compose

## Быстрый старт

### Запуск через Docker Compose

```bash
# Перейти в директорию deploy
cd deploy

# Сборка и запуск контейнеров
docker-compose up --build

# Запуск в фоновом режиме
docker-compose up -d --build

# Просмотр логов
docker-compose logs -f app

# Остановка контейнеров
docker-compose down
```

## API Endpoints

| Метод | URI | Описание | Статус |
|-------|-----|----------|--------|
| `GET` | `/health` | Проверка работоспособности | `200 OK` |
| `GET` | `/notes` | Список заметок (с пагинацией) | `200 OK` |
| `GET` | `/notes/:id` | Заметка по ID | `200 OK` / `404 Not Found` |
| `POST` | `/notes` | Создать заметку | `201 Created` |
| `PUT` | `/notes/:id` | Полное обновление заметки | `200 OK` / `404 Not Found` |
| `PATCH` | `/notes/:id` | Частичное обновление заметки | `200 OK` / `404 Not Found` |
| `DELETE` | `/notes/:id` | Удалить заметку (Soft Delete) | `204 No Content` / `404 Not Found` |

### Пагинация (GET /notes)

Параметры передаются через Query String:

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|-------------|----------|
| `page` | int | `1` | Номер страницы (min: 1) |
| `limit` | int | `10` | Кол-во записей на странице (min: 1, max: 100) |

### Обработка ошибок

| Код | Описание |
|-----|----------|
| `400 Bad Request` | Ошибка валидации (некорректный формат данных, пустой заголовок) |
| `404 Not Found` | Заметка не найдена или уже удалена |
| `500 Internal Server Error` | Внутренняя ошибка сервера |

## Структура проекта

```
.
├── deploy/
│   ├── .dockerignore      # Исключения для Docker
│   ├── .env.example       # Шаблон переменных окружения
│   ├── docker-compose.yml # Оркестрация сервисов
│   └── Dockerfile         # Образ приложения
├── internal/
│   ├── apperrors/         # Кастомные ошибки
│   ├── config/            # Конфигурация приложения
│   ├── database/          # Подключение к БД и миграции
│   ├── notes/
│   │   ├── handler/       # HTTP обработчики
│   │   ├── models/        # Модель Note (GORM)
│   │   ├── repository/    # Репозиторий (CRUD, интерфейс)
│   │   ├── routes/        # Маршруты Gin
│   │   ├── service/       # Бизнес-логика (интерфейсы)
│   │   └── validator/     # Валидация данных
├── .dockerignore
├── .env                   # Переменные окружения (не в git)
├── .gitignore
├── go.mod                 # Go модуль
├── go.sum                 # Go зависимости
├── main.go                # Точка входа
└── README.md              # Документация
```

## База данных

### Модель Note

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | UUID | Первичный ключ (автоматическая генерация) |
| `title` | VARCHAR(200) | Заголовок заметки (обязательное, не пустое) |
| `content` | TEXT | Текст заметки (необязательное) |
| `created_at` | TIMESTAMP WITH TIME ZONE | Дата создания |
| `updated_at` | TIMESTAMP WITH TIME ZONE | Дата последнего обновления |
| `deleted_at` | TIMESTAMP WITH TIME ZONE | Soft delete (NULL = активна) |

### Миграции

Схема базы данных создаётся автоматически через GORM AutoMigrate при каждом запуске приложения.

## Переменные окружения

| Переменная | По умолчанию | Описание |
|-------------|-------------|----------|
| `DB_USER` | `nix` | Пользователь PostgreSQL |
| `DB_PASSWORD` | `sews` | Пароль PostgreSQL |
| `DB_NAME` | `Web_Labs` | Имя базы данных |
| `DB_HOST` | `localhost` | Хост базы данных |
| `DB_PORT` | `5432` | Порт базы данных |
| `PORT` | `4200` | Порт приложения |

## Сервисы Docker Compose

| Сервис | Описание | Порт |
|--------|----------|------|
| `web_labs_db` | PostgreSQL 16 | `5432` |
| `web_labs_app` | Go-приложение | `4200` |
| `web_labs_pgadmin` | pgAdmin 4 | `5050` |

pgAdmin доступен по адресу: http://localhost:5050
- **Email:** `admin@admin.com`
- **Пароль:** `admin`

## Зависимости

- [Gin](https://github.com/gin-gonic/gin) — HTTP фреймворк
- [GORM](https://gorm.io/) — ORM для Go
- [UUID](https://github.com/google/uuid) — Генерация UUID
