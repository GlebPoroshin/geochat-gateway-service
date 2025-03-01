FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY geochat-shared /app/geochat-shared

COPY geochat-gateway-service /app/geochat-gateway-service

WORKDIR /app/geochat-gateway-service

RUN go mod download

RUN go build -o /gateway-service ./cmd/gateway-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /gateway-service .
EXPOSE 8080
CMD ["./gateway-service"] 