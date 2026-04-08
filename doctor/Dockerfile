FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY utils ./utils
COPY doctor ./doctor

RUN go build -o doctor-service ./doctor/cmd/doctor

FROM debian:bookworm-slim

WORKDIR /root/
COPY --from=builder /app/doctor-service .

EXPOSE 8081

CMD ["./doctor-service"]
