package repository

import (
	"errors"

	"github.com/Akendo/assigment1/doctor/internal/model"
)

var ErrDoctorNotFound = errors.New("doctor not found")
var ErrDoctorEmailAlreadyExists = errors.New("doctor email already exists")

type DoctorRepository interface {
	Create(doctor *model.Doctor) error
	GetByID(id string) (*model.Doctor, error)
	List() ([]*model.Doctor, error)
}
