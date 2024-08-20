# Этап 1: Сборка бинарного файла
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем зависимости Go
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код приложения
COPY . .

# Сборка приложения
RUN go build -o Calendar_main ./cmd/calendar/main.go


# Этап 2: Создание минимального образа
FROM alpine:3.18

WORKDIR /app

# Копируем скомпилированный бинарник из предыдущего этапа
COPY --from=builder /app/Calendar_main .

# Указываем команду для запуска сервиса
CMD ["sh", "-c", "sleep 15 && ./Calendar_main"]
