package cache

import (
	"context"
	"log"
	"time"

	"github.com/Akendo/assigment1/appointment/internal/model"
)

const (
	appointmentsListKey   = "appointments:list"
	cacheOperationTimeout = 2 * time.Second
)

type appointmentService interface {
	CreateAppointment(title, description, doctorID string) (*model.Appointment, error)
	GetAppointment(id string) (*model.Appointment, error)
	ListAppointments() ([]*model.Appointment, error)
	UpdateStatus(id string, status model.Status) (*model.Appointment, error)
}

type AppointmentService struct {
	next   appointmentService
	cache  CacheRepository
	ttl    time.Duration
	logger *log.Logger
}

func NewAppointmentService(next appointmentService, cache CacheRepository, ttl time.Duration, logger *log.Logger) *AppointmentService {
	if logger == nil {
		logger = log.Default()
	}

	return &AppointmentService{
		next:   next,
		cache:  cache,
		ttl:    ttl,
		logger: logger,
	}
}

func (s *AppointmentService) CreateAppointment(title, description, doctorID string) (*model.Appointment, error) {
	appointment, err := s.next.CreateAppointment(title, description, doctorID)
	if err != nil {
		return nil, err
	}

	s.delete(appointmentsListKey)

	return appointment, nil
}

func (s *AppointmentService) GetAppointment(id string) (*model.Appointment, error) {
	if s.cache != nil {
		key := appointmentByIDKey(id)
		var cached model.Appointment

		ctx, cancel := s.cacheContext()
		found, err := s.cache.Get(ctx, key, &cached)
		cancel()
		if err != nil {
			s.logger.Printf("appointment cache get %q: %v", key, err)
		} else if found {
			s.logger.Printf("appointment cache hit %q", key)
			return &cached, nil
		}
	}

	appointment, err := s.next.GetAppointment(id)
	if err != nil {
		return nil, err
	}

	s.set(appointmentByIDKey(id), appointment)

	return appointment, nil
}

func (s *AppointmentService) ListAppointments() ([]*model.Appointment, error) {
	if s.cache != nil {
		var cached []*model.Appointment

		ctx, cancel := s.cacheContext()
		found, err := s.cache.Get(ctx, appointmentsListKey, &cached)
		cancel()
		if err != nil {
			s.logger.Printf("appointment cache get %q: %v", appointmentsListKey, err)
		} else if found {
			s.logger.Printf("appointment cache hit %q", appointmentsListKey)
			return cached, nil
		}
	}

	appointments, err := s.next.ListAppointments()
	if err != nil {
		return nil, err
	}

	s.set(appointmentsListKey, appointments)

	return appointments, nil
}

func (s *AppointmentService) UpdateStatus(id string, status model.Status) (*model.Appointment, error) {
	appointment, err := s.next.UpdateStatus(id, status)
	if err != nil {
		return nil, err
	}

	s.set(appointmentByIDKey(id), appointment)
	s.delete(appointmentsListKey)

	return appointment, nil
}

func (s *AppointmentService) set(key string, value any) {
	if s.cache == nil {
		return
	}

	ctx, cancel := s.cacheContext()
	defer cancel()

	if err := s.cache.Set(ctx, key, value, s.ttl); err != nil {
		s.logger.Printf("appointment cache set %q: %v", key, err)
	}
}

func (s *AppointmentService) delete(key string) {
	if s.cache == nil {
		return
	}

	ctx, cancel := s.cacheContext()
	defer cancel()

	if err := s.cache.Delete(ctx, key); err != nil {
		s.logger.Printf("appointment cache delete %q: %v", key, err)
	}
}

func (s *AppointmentService) cacheContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), cacheOperationTimeout)
}

func appointmentByIDKey(id string) string {
	return "appointment:" + id
}
