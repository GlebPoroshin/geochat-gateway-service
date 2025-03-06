FROM golang:1.24-alpine AS builder

ENV GOPROXY=direct
ENV GO111MODULE=on

RUN apk add --no-cache git

WORKDIR /app

COPY geochat-gateway-service /app/geochat-gateway-service

WORKDIR /app/geochat-gateway-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/gateway-service ./cmd/gateway-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/gateway-service /app/gateway-service
EXPOSE 8080
CMD ["/app/gateway-service"]