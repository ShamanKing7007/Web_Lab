# Этап сборки
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Копируем файлы модуля
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код и миграции
COPY . .

# Собираем приложение
RUN go build -o main .

# Финальный этап
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/main /app/main

USER appuser

EXPOSE 4200

CMD ["/app/main"]
