FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY notification-service ./notification-service

RUN go build -o notification-service ./notification-service/cmd/notification-service

FROM debian:bookworm-slim

WORKDIR /root/
COPY --from=builder /app/notification-service .

CMD ["./notification-service"]
