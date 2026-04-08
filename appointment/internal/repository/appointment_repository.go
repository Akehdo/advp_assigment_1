package repository

import (
	"errors"

	"github.com/Akendo/assigment1/appointment/internal/model"
)

var ErrAppointmentNotFound = errors.New("appointment not found")

type AppointmentRepository interface {
	Create(appointment *model.Appointment) error
	GetByID(id string) (*model.Appointment, error)
	List() ([]*model.Appointment, error)
	UpdateStatus(id string, status model.Status) error
}
