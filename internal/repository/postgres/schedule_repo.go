package postgres

import (
	"database/sql"
	"medlink/internal/domain"
)

type ScheduleRepository struct{ db *sql.DB }

func NewScheduleRepo(db *sql.DB) domain.ScheduleRepository {
	return &ScheduleRepository{db}
}

func (r *ScheduleRepository) Create(s *domain.Schedule) error {
	const q = `
INSERT INTO schedules
  (doctor_id, service_id, start_time, end_time, color, visible)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id`
	return r.db.QueryRow(
		q,
		s.DoctorID, s.ServiceID,
		s.StartTime, s.EndTime,
		s.Color, s.Visible,
	).Scan(&s.ID)
}

func (r *ScheduleRepository) GetByID(id int) (*domain.Schedule, error) {
	const q = `
SELECT id, doctor_id, service_id, start_time, end_time, color, visible
FROM schedules
WHERE id = $1`
	var s domain.Schedule
	if err := r.db.QueryRow(q, id).Scan(
		&s.ID, &s.DoctorID, &s.ServiceID,
		&s.StartTime, &s.EndTime,
		&s.Color, &s.Visible,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ScheduleRepository) ListByDoctor(doctorID int) ([]domain.Schedule, error) {
	const q = `
SELECT id, doctor_id, service_id, start_time, end_time, color, visible
FROM schedules WHERE doctor_id = $1`
	rows, err := r.db.Query(q, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Schedule
	for rows.Next() {
		var s domain.Schedule
		if err := rows.Scan(
			&s.ID, &s.DoctorID, &s.ServiceID,
			&s.StartTime, &s.EndTime,
			&s.Color, &s.Visible,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ScheduleRepository) ToggleVisible(id int, visible bool) error {
	_, err := r.db.Exec(
		`UPDATE schedules SET visible=$2 WHERE id=$1`,
		id, visible,
	)
	return err
}
