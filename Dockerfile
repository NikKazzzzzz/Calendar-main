# Используем официальный образ Go для сборки бинарника
FROM golang:1.22 as builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

# Копируем исходный код и конфигурации
COPY cmd/ .
COPY config/local.yaml ./config/local.yaml

# Сборка бинарного файла
RUN go mod tidy
RUN go build -o calendar-main

# Создаем минимальный образ для запуска
FROM alpine:3.18

WORKDIR /app

# Копируем бинарник и конфиг из билдера
COPY --from=builder /app/calendar-main .
COPY --from=builder /app/config/local.yaml ./config/local.yaml

# Команда для запуска приложения
CMD ["./calendar-main"]
