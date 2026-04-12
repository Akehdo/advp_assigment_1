package usecase

import (
	"errors"

	"github.com/Akendo/assigment1/appointment/internal/model"
	"github.com/Akendo/assigment1/appointment/internal/repository"
)

var ErrDoctorServiceUnavailable = errors.New("doctor service is unavailable")
var ErrDoctorNotFound = errors.New("doctor not found")

type DoctorGateway interface {
	Exists(id string) (bool, error)
}

type AppointmentService struct {
	repo          repository.AppointmentRepository
	doctorGateway DoctorGateway
}

func NewAppointmentService(repo repository.AppointmentRepository, doctorGateway DoctorGateway) *AppointmentService {
	return &AppointmentService{
		repo:          repo,
		doctorGateway: doctorGateway,
	}
}

func (s *AppointmentService) CreateAppointment(title, description, doctorID string) (*model.Appointment, error) {
	if s.doctorGateway == nil {
		return nil, ErrDoctorServiceUnavailable
	}

	exists, err := s.doctorGateway.Exists(doctorID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDoctorNotFound
	}

	appointment, err := model.NewAppointment(title, description, doctorID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(appointment); err != nil {
		return nil, err
	}

	return appointment, nil
}

func (s *AppointmentService) GetAppointment(id string) (*model.Appointment, error) {
	return s.repo.GetByID(id)
}

func (s *AppointmentService) ListAppointments() ([]*model.Appointment, error) {
	return s.repo.List()
}

func (s *AppointmentService) UpdateStatus(id string, status model.Status) (*model.Appointment, error) {
	appointment, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := appointment.UpdateStatus(status); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateStatus(id, appointment.Status); err != nil {
		return nil, err
	}

	return appointment, nil
}
