# Используем образ Go для сборки
FROM golang:1.25-alpine3.21 AS builder

ENV GOLANG_VERSION 1.25.1

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем только файлы с зависимостями для кэширования
COPY go.mod go.sum ./

# Устанавливаем зависимости и инструменты
RUN go mod download

# Копируем исходный код
COPY . .

# Упрощаем и объединяем go мод команды
RUN go mod tidy && \
    go mod vendor

# Компиляция приложения
RUN go build -o ./bin/shorturl ./cmd/main.go

# Финальный образ с минимальным размером
FROM alpine:latest

ENV GOLANG_VERSION 1.25.1

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем необходимые файлы из билд-образа
COPY --from=builder /app/bin/shorturl ./bin/shorturl
COPY --from=builder /app/vendor ./vendor
COPY --from=builder /app/templates ./templates

# Экспортируем порт
EXPOSE 8082

# Запускаем приложение
CMD ["./bin/shorturl"]