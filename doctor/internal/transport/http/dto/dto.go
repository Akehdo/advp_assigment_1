package dto

import "github.com/Akendo/assigment1/doctor/internal/model"

type CreateDoctorRequest struct {
	FullName       string `json:"full_name" binding:"required"`
	Specialization string `json:"specialization"`
	Email          string `json:"email" binding:"required,email"`
}

type DoctorResponse struct {
	ID             string `json:"id"`
	FullName       string `json:"full_name"`
	Specialization string `json:"specialization"`
	Email          string `json:"email"`
}

func NewDoctorResponse(doctor *model.Doctor) DoctorResponse {
	return DoctorResponse{
		ID:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}
}

func NewDoctorListResponse(doctors []*model.Doctor) []DoctorResponse {
	response := make([]DoctorResponse, 0, len(doctors))
	for _, doctor := range doctors {
		response = append(response, NewDoctorResponse(doctor))
	}

	return response
}
