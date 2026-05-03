package event

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Akendo/assigment1/appointment/internal/model"
	"github.com/Akendo/assigment1/pkg/events"
	"github.com/nats-io/nats.go"
)

type Publisher interface {
	PublishAppointmentCreated(appointment *model.Appointment)
	PublishAppointmentStatusUpdated(appointmentID string, oldStatus, newStatus model.Status)
}

type NoopPublisher struct{}

func NewNoopPublisher() NoopPublisher {
	return NoopPublisher{}
}

func (NoopPublisher) PublishAppointmentCreated(*model.Appointment) {}

func (NoopPublisher) PublishAppointmentStatusUpdated(string, model.Status, model.Status) {}

type NATSPublisher struct {
	conn   *nats.Conn
	logger *log.Logger
}

func NewNATSPublisher(conn *nats.Conn, logger *log.Logger) *NATSPublisher {
	if logger == nil {
		logger = log.Default()
	}

	return &NATSPublisher{
		conn:   conn,
		logger: logger,
	}
}

func (p *NATSPublisher) PublishAppointmentCreated(appointment *model.Appointment) {
	if appointment == nil || p.conn == nil {
		return
	}

	payload := events.AppointmentCreatedEvent{
		EventType:  events.SubjectAppointmentCreated,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		ID:         appointment.ID,
		Title:      appointment.Title,
		DoctorID:   appointment.DoctorID,
		Status:     string(appointment.Status),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		p.logger.Printf("failed to marshal %s event for appointment %s: %v", events.SubjectAppointmentCreated, appointment.ID, err)
		return
	}

	if err := p.conn.Publish(events.SubjectAppointmentCreated, data); err != nil {
		p.logger.Printf("failed to publish %s event for appointment %s: %v", events.SubjectAppointmentCreated, appointment.ID, err)
	}
}

func (p *NATSPublisher) PublishAppointmentStatusUpdated(appointmentID string, oldStatus, newStatus model.Status) {
	if appointmentID == "" || p.conn == nil {
		return
	}

	payload := events.AppointmentStatusUpdatedEvent{
		EventType:  events.SubjectAppointmentStatusUpdated,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		ID:         appointmentID,
		OldStatus:  string(oldStatus),
		NewStatus:  string(newStatus),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		p.logger.Printf("failed to marshal %s event for appointment %s: %v", events.SubjectAppointmentStatusUpdated, appointmentID, err)
		return
	}

	if err := p.conn.Publish(events.SubjectAppointmentStatusUpdated, data); err != nil {
		p.logger.Printf("failed to publish %s event for appointment %s: %v", events.SubjectAppointmentStatusUpdated, appointmentID, err)
	}
}
