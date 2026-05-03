package event

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Akendo/assigment1/doctor/internal/model"
	"github.com/Akendo/assigment1/pkg/events"
	"github.com/nats-io/nats.go"
)

type Publisher interface {
	PublishDoctorCreated(doctor *model.Doctor)
}

type NoopPublisher struct{}

func NewNoopPublisher() NoopPublisher {
	return NoopPublisher{}
}

func (NoopPublisher) PublishDoctorCreated(*model.Doctor) {}

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

func (p *NATSPublisher) PublishDoctorCreated(doctor *model.Doctor) {
	if doctor == nil || p.conn == nil {
		return
	}

	payload := events.DoctorCreatedEvent{
		EventType:      events.SubjectDoctorCreated,
		OccurredAt:     time.Now().UTC().Format(time.RFC3339),
		ID:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		p.logger.Printf("failed to marshal %s event for doctor %s: %v", events.SubjectDoctorCreated, doctor.ID, err)
		return
	}

	if err := p.conn.Publish(events.SubjectDoctorCreated, data); err != nil {
		p.logger.Printf("failed to publish %s event for doctor %s: %v", events.SubjectDoctorCreated, doctor.ID, err)
	}
}
