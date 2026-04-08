package dto

import (
	"time"

	"github.com/Akendo/assigment1/appointment/internal/model"
)

type CreateAppointmentRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	DoctorID    string `json:"doctor_id" binding:"required,uuid"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=new in_progress done"`
}

type AppointmentResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DoctorID    string    `json:"doctor_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewAppointmentResponse(appointment *model.Appointment) AppointmentResponse {
	return AppointmentResponse{
		ID:          appointment.ID,
		Title:       appointment.Title,
		Description: appointment.Description,
		DoctorID:    appointment.DoctorID,
		Status:      string(appointment.Status),
		CreatedAt:   appointment.CreatedAt,
		UpdatedAt:   appointment.UpdatedAt,
	}
}

func NewAppointmentListResponse(appointments []*model.Appointment) []AppointmentResponse {
	response := make([]AppointmentResponse, 0, len(appointments))
	for _, appointment := range appointments {
		response = append(response, NewAppointmentResponse(appointment))
	}

	return response
}
