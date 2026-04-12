package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/Akendo/assigment1/doctor/internal/model"
	"github.com/Akendo/assigment1/doctor/internal/repository"
	doctorpb "github.com/Akendo/assigment1/doctor/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type doctorService interface {
	CreateDoctor(fullName, specialization, email string) (*model.Doctor, error)
	GetDoctor(id string) (*model.Doctor, error)
	ListDoctors() ([]*model.Doctor, error)
}

type Handler struct {
	doctorpb.UnimplementedDoctorServiceServer
	service doctorService
}

func NewHandler(service doctorService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateDoctor(ctx context.Context, req *doctorpb.CreateDoctorRequest) (*doctorpb.DoctorResponse, error) {
	if strings.TrimSpace(req.GetFullName()) == "" {
		return nil, status.Error(codes.InvalidArgument, "full_name is required")
	}

	if strings.TrimSpace(req.GetEmail()) == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	doctor, err := h.service.CreateDoctor(req.GetFullName(), req.GetSpecialization(), req.GetEmail())
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDoctorEmailAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return toDoctorResponse(doctor), nil
}

func (h *Handler) GetDoctor(ctx context.Context, req *doctorpb.GetDoctorRequest) (*doctorpb.DoctorResponse, error) {
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	doctor, err := h.service.GetDoctor(req.GetId())
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDoctorNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return toDoctorResponse(doctor), nil
}

func (h *Handler) ListDoctors(ctx context.Context, req *doctorpb.ListDoctorsRequest) (*doctorpb.ListDoctorsResponse, error) {
	doctors, err := h.service.ListDoctors()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := make([]*doctorpb.DoctorResponse, 0, len(doctors))
	for _, doctor := range doctors {
		response = append(response, toDoctorResponse(doctor))
	}

	return &doctorpb.ListDoctorsResponse{
		Doctors: response,
	}, nil
}

func toDoctorResponse(doctor *model.Doctor) *doctorpb.DoctorResponse {
	return &doctorpb.DoctorResponse{
		Id:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}
}
