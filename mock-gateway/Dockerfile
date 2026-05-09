FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY mock-gateway ./mock-gateway

RUN go build -o mock-gateway ./mock-gateway/cmd/mock-gateway

FROM debian:bookworm-slim

WORKDIR /root/
COPY --from=builder /app/mock-gateway .

EXPOSE 8080

CMD ["./mock-gateway"]
