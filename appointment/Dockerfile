FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY utils ./utils
COPY appointment ./appointment

RUN go build -o appointment-service ./appointment/cmd/appointment

FROM debian:bookworm-slim

WORKDIR /root/
COPY --from=builder /app/appointment-service .

EXPOSE 8082

CMD ["./appointment-service"]
