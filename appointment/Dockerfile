FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY utils ./utils
COPY doctor/proto ./doctor/proto
COPY appointment ./appointment

RUN go build -o appointment-service ./appointment/cmd/appointment

FROM debian:bookworm-slim

WORKDIR /root/
COPY --from=builder /app/appointment-service .

EXPOSE 50052

CMD ["./appointment-service"]
