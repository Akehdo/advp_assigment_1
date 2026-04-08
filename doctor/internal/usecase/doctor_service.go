package usecase

import (
	"github.com/Akendo/assigment1/doctor/internal/model"
	"github.com/Akendo/assigment1/doctor/internal/repository"
)

type DoctorService struct {
	repo repository.DoctorRepository
}

func NewDoctorService(repo repository.DoctorRepository) *DoctorService {
	return &DoctorService{repo: repo}
}

func (s *DoctorService) CreateDoctor(fullName, specialization, email string) (*model.Doctor, error) {
	doctor, err := model.NewDoctor(fullName, specialization, email)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(doctor); err != nil {
		return nil, err
	}

	return doctor, nil
}

func (s *DoctorService) GetDoctor(id string) (*model.Doctor, error) {
	return s.repo.GetByID(id)
}

func (s *DoctorService) ListDoctors() ([]*model.Doctor, error) {
	return s.repo.List()
}
