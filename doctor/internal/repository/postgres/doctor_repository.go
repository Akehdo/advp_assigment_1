package postgres

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/Akendo/assigment1/doctor/internal/model"
	"github.com/Akendo/assigment1/doctor/internal/repository"
)

type DoctorRepository struct {
	db *sql.DB
}

func NewDoctorRepository(db *sql.DB) *DoctorRepository {
	return &DoctorRepository{db: db}
}

func (r *DoctorRepository) Create(doctor *model.Doctor) error {
	const query = `
		INSERT INTO doctors (id, full_name, specialization, email)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(query,
		doctor.ID,
		doctor.FullName,
		doctor.Specialization,
		doctor.Email,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return repository.ErrDoctorEmailAlreadyExists
		}

		return err
	}

	return nil
}

func (r *DoctorRepository) GetByID(id string) (*model.Doctor, error) {
	const query = `
		SELECT id, full_name, specialization, email
		FROM doctors
		WHERE id = $1
	`

	var doctor model.Doctor
	err := r.db.QueryRow(query, id).Scan(
		&doctor.ID,
		&doctor.FullName,
		&doctor.Specialization,
		&doctor.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrDoctorNotFound
		}

		return nil, err
	}

	return &doctor, nil
}

func (r *DoctorRepository) List() ([]*model.Doctor, error) {
	const query = `
		SELECT id, full_name, specialization, email
		FROM doctors
		ORDER BY full_name ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	doctors := make([]*model.Doctor, 0)
	for rows.Next() {
		var doctor model.Doctor
		if err := rows.Scan(
			&doctor.ID,
			&doctor.FullName,
			&doctor.Specialization,
			&doctor.Email,
		); err != nil {
			return nil, err
		}

		doctors = append(doctors, &doctor)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return doctors, nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key value")
}
