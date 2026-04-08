package model

import (
	"time"

	"github.com/Akendo/assigment1/utils/uuid"
)

type Appointment struct {
	ID          string
	Title       string
	Description string
	DoctorID    string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewAppointment(title, description, doctorID string) (*Appointment, error) {
	id, err := uuid.NewString()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &Appointment{
		ID:          id,
		Title:       title,
		Description: description,
		DoctorID:    doctorID,
		Status:      StatusNew,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (a *Appointment) UpdateStatus(status Status) error {
	if !status.IsValid() {
		return ErrInvalidStatus
	}
	if a.Status == StatusDone && status == StatusNew {
		return ErrStatusTransitionNotAllowed
	}

	a.Status = status
	a.UpdatedAt = time.Now()
	return nil
}
