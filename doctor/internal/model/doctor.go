package model

import "github.com/Akendo/assigment1/utils/uuid"

type Doctor struct {
	ID             string
	FullName       string
	Specialization string
	Email          string
}

func NewDoctor(fullName, specialization, email string) (*Doctor, error) {
	id, err := uuid.NewString()
	if err != nil {
		return nil, err
	}

	return &Doctor{
		ID:             id,
		FullName:       fullName,
		Specialization: specialization,
		Email:          email,
	}, nil
}
