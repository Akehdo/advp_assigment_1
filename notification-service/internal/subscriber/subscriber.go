package subscriber

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Akendo/assigment1/pkg/events"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	conn    *nats.Conn
	logger  *log.Logger
	encoder *json.Encoder
	mu      sync.Mutex
}

func New(conn *nats.Conn, logger *log.Logger) *Subscriber {
	if logger == nil {
		logger = log.Default()
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)

	return &Subscriber{
		conn:    conn,
		logger:  logger,
		encoder: encoder,
	}
}

func (s *Subscriber) SubscribeAll() error {
	subjects := []string{
		events.SubjectDoctorCreated,
		events.SubjectAppointmentCreated,
		events.SubjectAppointmentStatusUpdated,
	}

	for _, subject := range subjects {
		if _, err := s.conn.Subscribe(subject, s.handleMessage); err != nil {
			return err
		}
	}

	return s.conn.Flush()
}

func (s *Subscriber) handleMessage(msg *nats.Msg) {
	var event map[string]any
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Printf("failed to decode message on %s: %v", msg.Subject, err)
		return
	}

	entry := events.NotificationLogLine{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Subject: msg.Subject,
		Event:   event,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.encoder.Encode(entry); err != nil {
		s.logger.Printf("failed to write notification log for %s: %v", msg.Subject, err)
	}
}
