package postgres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Akendo/assigment1/appointment/internal/model"
	"github.com/Akendo/assigment1/appointment/internal/repository"
)

type AppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Create(appointment *model.Appointment) error {
	const query = `
		INSERT INTO appointments (id, title, description, doctor_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(query,
		appointment.ID,
		appointment.Title,
		appointment.Description,
		appointment.DoctorID,
		appointment.Status,
		appointment.CreatedAt,
		appointment.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *AppointmentRepository) GetByID(id string) (*model.Appointment, error) {
	const query = `
		SELECT id, title, description, doctor_id, status, created_at, updated_at
		FROM appointments
		WHERE id = $1
	`

	var appointment model.Appointment
	err := r.db.QueryRow(query, id).Scan(
		&appointment.ID,
		&appointment.Title,
		&appointment.Description,
		&appointment.DoctorID,
		&appointment.Status,
		&appointment.CreatedAt,
		&appointment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrAppointmentNotFound
		}

		return nil, err
	}

	return &appointment, nil
}

func (r *AppointmentRepository) List() ([]*model.Appointment, error) {
	const query = `
		SELECT id, title, description, doctor_id, status, created_at, updated_at
		FROM appointments
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appointments := make([]*model.Appointment, 0)
	for rows.Next() {
		var appointment model.Appointment
		if err := rows.Scan(
			&appointment.ID,
			&appointment.Title,
			&appointment.Description,
			&appointment.DoctorID,
			&appointment.Status,
			&appointment.CreatedAt,
			&appointment.UpdatedAt,
		); err != nil {
			return nil, err
		}

		appointments = append(appointments, &appointment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return appointments, nil
}

func (r *AppointmentRepository) UpdateStatus(id string, status model.Status) error {
	const query = `
		UPDATE appointments
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id, status, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return repository.ErrAppointmentNotFound
	}

	return nil
}
