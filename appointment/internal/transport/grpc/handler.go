package grpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Akendo/assigment1/appointment/internal/model"
	"github.com/Akendo/assigment1/appointment/internal/repository"
	"github.com/Akendo/assigment1/appointment/internal/usecase"
	appointmentpb "github.com/Akendo/assigment1/appointment/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type appointmentService interface {
	CreateAppointment(title, description, doctorID string) (*model.Appointment, error)
	GetAppointment(id string) (*model.Appointment, error)
	ListAppointments() ([]*model.Appointment, error)
	UpdateStatus(id string, status model.Status) (*model.Appointment, error)
}

type Handler struct {
	appointmentpb.UnimplementedAppointmentServiceServer
	service appointmentService
}

func NewHandler(service appointmentService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateAppointment(ctx context.Context, req *appointmentpb.CreateAppointmentRequest) (*appointmentpb.AppointmentResponse, error) {
	if strings.TrimSpace(req.GetTitle()) == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	if strings.TrimSpace(req.GetDoctorId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "doctor_id is required")
	}

	appointment, err := h.service.CreateAppointment(req.GetTitle(), req.GetDescription(), req.GetDoctorId())
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrDoctorNotFound):
			return nil, status.Error(codes.FailedPrecondition, "doctor does not exist")
		case errors.Is(err, usecase.ErrDoctorServiceUnavailable):
			return nil, status.Error(codes.Unavailable, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return toAppointmentResponse(appointment), nil
}

func (h *Handler) GetAppointment(ctx context.Context, req *appointmentpb.GetAppointmentRequest) (*appointmentpb.AppointmentResponse, error) {
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	appointment, err := h.service.GetAppointment(req.GetId())
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAppointmentNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return toAppointmentResponse(appointment), nil
}

func (h *Handler) ListAppointments(ctx context.Context, req *appointmentpb.ListAppointmentsRequest) (*appointmentpb.ListAppointmentsResponse, error) {
	appointments, err := h.service.ListAppointments()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := make([]*appointmentpb.AppointmentResponse, 0, len(appointments))
	for _, appointment := range appointments {
		response = append(response, toAppointmentResponse(appointment))
	}

	return &appointmentpb.ListAppointmentsResponse{
		Appointments: response,
	}, nil
}

func (h *Handler) UpdateAppointmentStatus(ctx context.Context, req *appointmentpb.UpdateStatusRequest) (*appointmentpb.AppointmentResponse, error) {
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if strings.TrimSpace(req.GetStatus()) == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	appointment, err := h.service.UpdateStatus(req.GetId(), model.Status(req.GetStatus()))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAppointmentNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, model.ErrInvalidStatus), errors.Is(err, model.ErrStatusTransitionNotAllowed):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return toAppointmentResponse(appointment), nil
}

func toAppointmentResponse(appointment *model.Appointment) *appointmentpb.AppointmentResponse {
	return &appointmentpb.AppointmentResponse{
		Id:          appointment.ID,
		Title:       appointment.Title,
		Description: appointment.Description,
		DoctorId:    appointment.DoctorID,
		Status:      string(appointment.Status),
		CreatedAt:   appointment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   appointment.UpdatedAt.Format(time.RFC3339),
	}
}
