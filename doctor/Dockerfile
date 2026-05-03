FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY utils ./utils
COPY doctor ./doctor

RUN go build -o doctor-service ./doctor/cmd/doctor

FROM debian:bookworm-slim

WORKDIR /root/
COPY --from=builder /app/doctor-service .
COPY --from=builder /app/doctor/migrations ./migrations

EXPOSE 50051

CMD ["./doctor-service"]
