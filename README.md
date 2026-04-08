# Doctor Appointment API

This project contains two REST microservices 

- `doctor` service
- `appointment` service

Each service has its own database and communicates over HTTP.

### Doctor Service

Base URL:

```text
http://localhost:8081
```

Endpoints:

- `GET /health`
- `POST /doctors`
- `GET /doctors/:id`
- `GET /doctors`


### Appointment Service

Base URL:

```text
http://localhost:8082
```

Endpoints:

- `GET /health`
- `POST /appointments`
- `GET /appointments/:id`
- `GET /appointments`
- `PATCH /appointments/:id/status`


## Run with Docker

Start everything:

```bash
docker compose up --build
```

This starts:

- `doctor-service` on port `8081`
- `appointment-service` on port `8082`

## API Examples

### Create doctor

Request:

```http
POST /doctors
Content-Type: application/json
```

```json
{
  "full_name": "Dr. John Smith",
  "specialization": "Cardiology",
  "email": "john.smith@example.com"
}
```

Example:

```bash
curl -X POST http://localhost:8081/doctors \
  -H "Content-Type: application/json" \
  -d "{\"full_name\":\"Dr. John Smith\",\"specialization\":\"Cardiology\",\"email\":\"john.smith@example.com\"}"
```

### Get doctor by ID

```bash
curl http://localhost:8081/doctors/{doctor_id}
```

### List doctors

```bash
curl http://localhost:8081/doctors
```

### Create appointment

Request:

```http
POST /appointments
Content-Type: application/json
```

```json
{
  "title": "Initial consultation",
  "description": "First visit for general checkup",
  "doctor_id": "PUT_DOCTOR_ID_HERE"
}
```

Example:

```bash
curl -X POST http://localhost:8082/appointments \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"Initial consultation\",\"description\":\"First visit for general checkup\",\"doctor_id\":\"PUT_DOCTOR_ID_HERE\"}"
```

### Get appointment by ID

```bash
curl http://localhost:8082/appointments/{appointment_id}
```

### List appointments

```bash
curl http://localhost:8082/appointments
```

### Update appointment status

Request:

```http
PATCH /appointments/:id/status
Content-Type: application/json
```

```json
{
  "status": "in_progress"
}
```

Example:

```bash
curl -X PATCH http://localhost:8082/appointments/{appointment_id}/status \
  -H "Content-Type: application/json" \
  -d "{\"status\":\"done\"}"
```

