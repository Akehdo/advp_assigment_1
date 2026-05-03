package events

const (
	SubjectDoctorCreated            = "doctors.created"
	SubjectAppointmentCreated       = "appointments.created"
	SubjectAppointmentStatusUpdated = "appointments.status_updated"
)

type DoctorCreatedEvent struct {
	EventType      string `json:"event_type"`
	OccurredAt     string `json:"occurred_at"`
	ID             string `json:"id"`
	FullName       string `json:"full_name"`
	Specialization string `json:"specialization"`
	Email          string `json:"email"`
}

type AppointmentCreatedEvent struct {
	EventType  string `json:"event_type"`
	OccurredAt string `json:"occurred_at"`
	ID         string `json:"id"`
	Title      string `json:"title"`
	DoctorID   string `json:"doctor_id"`
	Status     string `json:"status"`
}

type AppointmentStatusUpdatedEvent struct {
	EventType  string `json:"event_type"`
	OccurredAt string `json:"occurred_at"`
	ID         string `json:"id"`
	OldStatus  string `json:"old_status"`
	NewStatus  string `json:"new_status"`
}

type NotificationLogLine struct {
	Time    string `json:"time"`
	Subject string `json:"subject"`
	Event   any    `json:"event"`
}
