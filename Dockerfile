FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем shared модуль
COPY geochat-shared /app/geochat-shared

# Копируем gateway-service
COPY geochat-gateway-service /app/geochat-gateway-service

WORKDIR /app/geochat-gateway-service

# Скачиваем зависимости
RUN go mod download

# Собираем приложение
RUN go build -o /gateway-service ./cmd/gateway-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /gateway-service .
EXPOSE 8080
CMD ["./gateway-service"] 