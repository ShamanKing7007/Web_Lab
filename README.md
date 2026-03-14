## Возможности

- **`/info`** — возвращает JSON с количеством дней до следующего Нового года

Пример ответа `/info`:
```json
{"days_until_new_year": 304}
```

## Запуск без Docker

```bash
# Загрузка зависимостей
go mod download

# Сборка приложения
go build -o main .

# Запуск сервера
go run main.go
```

Сервер запустится на порту **3000**.
- http://localhost:3000/info — информация о днях до Нового года

## Запуск через Docker
```bash
# Сборка и запуск контейнера
docker-compose up --build

# Остановка контейнера
docker-compose down
```


