# AP2 Assignment 3 - Message Queue and Database Migrations

This project extends the previous Medical Scheduling Platform from Assignment 2.

The system now includes:

- `doctor` service
- `appointment` service
- `notification-service`
- PostgreSQL for `doctor`
- PostgreSQL for `appointment`
- NATS as the message broker

The domain models, gRPC contracts, and use-case rules from Assignment 2 are preserved. The main changes in this assignment are:

- in-memory storage is replaced with PostgreSQL
- database schema is managed through SQL migration files
- successful write operations publish domain events to NATS
- `notification-service` subscribes to those events and logs them to stdout

## Broker Choice

This project uses `NATS (core)` as the message broker.

Reason for this choice:

- I had already worked with RabbitMQ before, so for this assignment I wanted to try `NATS` and compare a different messaging approach in practice
- `NATS` has a strong Go ecosystem and feels natural in a Go microservice project
- it is simpler to start and use locally than RabbitMQ
- it is a good fit for lightweight stateless event notifications
- it fully satisfies the assignment requirement of fire-and-forget pub/sub

If stronger delivery guarantees were needed in production, the current design could be improved with:

- `NATS JetStream` for durable delivery
- the `Outbox pattern` to avoid losing events between DB commit and publish

## Project Structure

```text
assigment-2/
├── doctor/
│   ├── cmd/doctor/
│   ├── internal/
│   │   ├── app/
│   │   ├── event/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── transport/grpc/
│   │   └── usecase/
│   ├── migrations/
│   └── proto/
├── appointment/
│   ├── cmd/appointment/
│   ├── internal/
│   │   ├── app/
│   │   ├── client/
│   │   ├── event/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── transport/grpc/
│   │   └── usecase/
│   ├── migrations/
│   └── proto/
├── notification-service/
│   ├── cmd/notification-service/
│   └── internal/
│       ├── app/
│       └── subscriber/
├── pkg/
│   ├── events/
│   └── messaging/
├── docker-compose.yml
├── go.mod
└── README.md
```

## Clean Architecture Notes

- `internal/model` contains domain entities and business rules
- `internal/repository` contains repository interfaces
- `internal/repository/postgres` contains PostgreSQL implementations
- `internal/usecase` contains application logic
- `internal/transport/grpc` contains thin gRPC handlers
- `internal/event` contains event publisher implementations
- `internal/app` wires the dependencies and starts each service

Infrastructure types do not leak into the domain model layer.

## Databases and Ownership

Each service owns its own database:

- `doctor` owns table `doctors`
- `appointment` owns table `appointments`

There is no shared table access across services.

## Environment Variables

### Doctor service

- `DATABASE_URL` or `DB_DSN`
- fallback variables also supported:
  - `DB_HOST`
  - `DB_PORT`
  - `DB_NAME`
  - `DB_USER`
  - `DB_PASSWORD`
- `NATS_URL`

### Appointment service

- `DATABASE_URL` or `DB_DSN`
- fallback variables also supported:
  - `DB_HOST`
  - `DB_PORT`
  - `DB_NAME`
  - `DB_USER`
  - `DB_PASSWORD`
- `DOCTOR_SERVICE_ADDR`
- `NATS_URL`

### Notification service

- `NATS_URL`

### Default values used in Docker Compose

```text
DOCTOR_DB_NAME=doctor_db
DOCTOR_DB_USER=doctor_user
DOCTOR_DB_PASSWORD=doctor_pass
DOCTOR_DB_PORT=5433

APPOINTMENT_DB_NAME=appointment_db
APPOINTMENT_DB_USER=appointment_user
APPOINTMENT_DB_PASSWORD=appointment_pass
APPOINTMENT_DB_PORT=5434

DOCTOR_GRPC_PORT=50051
APPOINTMENT_GRPC_PORT=50052
DOCTOR_SERVICE_ADDR=doctor:50051

NATS_URL=nats://nats:4222
NATS_PORT=4222
NATS_MONITORING_PORT=8222
```

## Infrastructure Setup

### Option 1. Start everything with Docker Compose

Docker Compose reads variables from the root `.env` file.

```powershell
docker compose up --build
```

This starts:

- `doctor-db` on host port `5433`
- `appointment-db` on host port `5434`
- `NATS` on host port `4222`
- `NATS monitoring` on host port `8222`
- `doctor-service` on `localhost:50051`
- `appointment-service` on `localhost:50052`
- `notification-service`

If you want to share the configuration with another machine, use `.env.example` as the template and create a local `.env`.

### NATS Monitoring

After startup you can inspect NATS in the browser:

- [http://localhost:8222/varz](http://localhost:8222/varz)
- [http://localhost:8222/connz](http://localhost:8222/connz)
- [http://localhost:8222/subsz](http://localhost:8222/subsz)

## Service Startup Order

Recommended order:

1. PostgreSQL containers
2. NATS
3. `doctor` service
4. `appointment` service
5. `notification-service`

With Docker Compose this order is already coordinated by `depends_on` and database health checks.

Why this order:

- both gRPC services require their PostgreSQL instances before startup because migrations run at boot
- `appointment` depends on `doctor` for doctor existence checks over gRPC
- `notification-service` depends on NATS because it subscribes on startup

### Exact `go run .` commands

If services are started manually instead of Docker Compose, first set the required environment variables in the shell, then run each service from its main package directory.

Start `doctor` first:

```powershell
Set-Location doctor/cmd/doctor
go run .
```

Start `appointment` second:

```powershell
Set-Location appointment/cmd/appointment
go run .
```

Start `notification-service` third:

```powershell
Set-Location notification-service/cmd/notification-service
go run .
```

If you want to return to the repository root between commands:

```powershell
Set-Location ../..
```

## Migrations

Migration files are stored inside each service:

- `doctor/migrations/`
- `appointment/migrations/`

Current files:

- `doctor/migrations/000001_create_doctors.up.sql`
- `doctor/migrations/000001_create_doctors.down.sql`
- `appointment/migrations/000001_create_appointments.up.sql`
- `appointment/migrations/000001_create_appointments.down.sql`

### Automatic behavior

Migrations run automatically on service startup before the gRPC server starts accepting requests.

### Manual CLI examples

Doctor DB:

```powershell
migrate -path doctor/migrations -database "postgres://doctor_user:doctor_pass@localhost:5433/doctor_db?sslmode=disable" up
migrate -path doctor/migrations -database "postgres://doctor_user:doctor_pass@localhost:5433/doctor_db?sslmode=disable" down 1
```

Appointment DB:

```powershell
migrate -path appointment/migrations -database "postgres://appointment_user:appointment_pass@localhost:5434/appointment_db?sslmode=disable" up
migrate -path appointment/migrations -database "postgres://appointment_user:appointment_pass@localhost:5434/appointment_db?sslmode=disable" down 1
```

No raw DDL is executed from application code outside migration files.

## Database Schema

### `doctors`

```sql
CREATE TABLE doctors (
  id TEXT PRIMARY KEY,
  full_name TEXT NOT NULL,
  specialization TEXT NOT NULL DEFAULT '',
  email TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### `appointments`

```sql
CREATE TABLE appointments (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  doctor_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'new',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## Event Contract

All published events are JSON and include:

- `event_type`
- `occurred_at` in RFC3339 format
- entity-specific fields

### `doctors.created`

Publisher: `doctor` service  
Trigger: successful `CreateDoctor`

```json
{
  "event_type": "doctors.created",
  "occurred_at": "2026-05-03T12:05:10Z",
  "id": "doctor-id",
  "full_name": "Ruslan Kadirov",
  "specialization": "heart",
  "email": "ruslan.kadirov@example.com"
}
```

### `appointments.created`

Publisher: `appointment` service  
Trigger: successful `CreateAppointment`

```json
{
  "event_type": "appointments.created",
  "occurred_at": "2026-05-03T12:20:33Z",
  "id": "appointment-id",
  "title": "Initial cardiac consultation",
  "doctor_id": "doctor-id",
  "status": "new"
}
```

### `appointments.status_updated`

Publisher: `appointment` service  
Trigger: successful `UpdateAppointmentStatus`

```json
{
  "event_type": "appointments.status_updated",
  "occurred_at": "2026-05-03T12:21:01Z",
  "id": "appointment-id",
  "old_status": "new",
  "new_status": "done"
}
```

## Notification Service

`notification-service` has one responsibility:

- subscribe to `doctors.created`
- subscribe to `appointments.created`
- subscribe to `appointments.status_updated`
- log one structured JSON line to stdout for every received event

It does not:

- expose gRPC
- expose HTTP
- use a database
- call other services

### Example stdout log

```json
{"time":"2026-05-03T12:05:10Z","subject":"doctors.created","event":{"event_type":"doctors.created","occurred_at":"2026-05-03T12:05:10Z","id":"doctor-id","full_name":"Ruslan Kadirov","specialization":"heart","email":"ruslan.kadirov@example.com"}}
{"time":"2026-05-03T12:20:33Z","subject":"appointments.created","event":{"event_type":"appointments.created","occurred_at":"2026-05-03T12:20:33Z","id":"appointment-id","title":"Initial cardiac consultation","doctor_id":"doctor-id","status":"new"}}
{"time":"2026-05-03T12:21:01Z","subject":"appointments.status_updated","event":{"event_type":"appointments.status_updated","occurred_at":"2026-05-03T12:21:01Z","id":"appointment-id","old_status":"new","new_status":"done"}}
```

## gRPC Services

### Doctor Service RPCs

- `CreateDoctor`
- `GetDoctor`
- `ListDoctors`

### Appointment Service RPCs

- `CreateAppointment`
- `GetAppointment`
- `ListAppointments`
- `UpdateAppointmentStatus`

## Business Rules

### Doctor service

- `full_name` is required
- `email` is required
- `email` must be unique

### Appointment service

- `title` is required
- `doctor_id` is required
- referenced doctor must exist
- valid statuses are `new`, `in_progress`, `done`
- transition `done -> new` is forbidden

## Error Handling

### Database

- database unavailable on startup -> service exits with non-zero status
- runtime database failure -> gRPC returns `codes.Internal`
- duplicate doctor email -> `codes.AlreadyExists`
- row not found -> `codes.NotFound`

### Broker

- broker unavailable on startup for `doctor` or `appointment` -> service still starts and logs a warning
- broker publish failure during RPC -> request still succeeds, error is logged
- broker unavailable on startup for `notification-service` -> retries with exponential backoff, then exits with non-zero status

### Inter-service gRPC

- if `doctor` service is unreachable, `appointment` returns `codes.Unavailable`
- if doctor does not exist, `appointment` returns `codes.FailedPrecondition`

## Consistency Trade-offs

Event publishing in this project is `best-effort`.

This means:

- database write may succeed
- process may crash before event publish
- the event can be lost

This trade-off is acceptable for the assignment, but in production a more reliable solution would use:

- the `Outbox pattern`
- durable broker features such as `NATS JetStream`
- stronger delivery tracking and replay mechanisms

## NATS vs RabbitMQ

Two concrete differences:

1. Delivery model

- `NATS core` is lightweight pub/sub and fire-and-forget
- `RabbitMQ` provides durable queues and richer delivery guarantees

2. Operational complexity

- `NATS` is simpler to start and reason about locally
- `RabbitMQ` offers more delivery features but usually requires more setup and queue/exchange management

When I would choose each:

- choose `NATS` for simple internal event notifications with low operational overhead
- choose `RabbitMQ` when queue durability, acknowledgements, and stronger delivery guarantees matter

## Testing / Demo Flow

### 1. Create doctor

Request body:

```json
{
  "full_name": "Ruslan Kadirov",
  "specialization": "heart",
  "email": "ruslan.kadirov@example.com"
}
```

Expected result:

- doctor row inserted into `doctor_db`
- `doctors.created` event published
- `notification-service` logs a JSON line with subject `doctors.created`

### 2. Create appointment

Request body:

```json
{
  "title": "Initial cardiac consultation",
  "description": "First visit for heart check",
  "doctor_id": "PASTE_DOCTOR_ID_HERE"
}
```

Expected result:

- appointment row inserted into `appointment_db`
- `appointments.created` event published
- `notification-service` logs a JSON line with subject `appointments.created`

### 3. Update appointment status

Request body:

```json
{
  "id": "PASTE_APPOINTMENT_ID_HERE",
  "status": "done"
}
```

Expected result:

- appointment status updated in DB
- `appointments.status_updated` event published
- `notification-service` logs a JSON line with subject `appointments.status_updated`

## Tools Used

- Go
- gRPC
- PostgreSQL
- `github.com/golang-migrate/migrate/v4`
- NATS
- Docker Compose

## References

- [NATS documentation](https://docs.nats.io/)
- [NATS Go client](https://github.com/nats-io/nats.go)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [gRPC Go documentation](https://grpc.io/docs/languages/go/)
- [PostgreSQL documentation](https://www.postgresql.org/docs/)
