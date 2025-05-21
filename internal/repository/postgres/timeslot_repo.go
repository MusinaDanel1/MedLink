package postgres

import (
	"database/sql"
	"telemed/internal/domain"
	"time"
)

type TimeslotRepository struct {
	db *sql.DB
}

func NewTimeslotRepo(db *sql.DB) domain.TimeslotRepository {
	return &TimeslotRepository{db: db}
}

func (r *TimeslotRepository) GetOrCreate(scheduleID int, start, end time.Time) (*domain.TimeSlot, error) {
	// Check if timeslot exists
	var ts domain.TimeSlot
	err := r.db.QueryRow(
		"SELECT id, schedule_id, start_time, end_time, is_booked, created_at FROM timeslots WHERE schedule_id = $1 AND start_time = $2 AND end_time = $3",
		scheduleID, start, end,
	).Scan(&ts.ID, &ts.ScheduleID, &ts.StartTime, &ts.EndTime, &ts.IsBooked, &ts.CreatedAt)

	if err == nil {
		// Timeslot exists, return it
		return &ts, nil
	}

	// Timeslot doesn't exist, create it
	err = r.db.QueryRow(
		"INSERT INTO timeslots (schedule_id, start_time, end_time, is_booked) VALUES ($1, $2, $3, false) RETURNING id, schedule_id, start_time, end_time, is_booked, created_at",
		scheduleID, start, end,
	).Scan(&ts.ID, &ts.ScheduleID, &ts.StartTime, &ts.EndTime, &ts.IsBooked, &ts.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &ts, nil
}

func (r *TimeslotRepository) MarkBooked(id int, booked bool) error {
	_, err := r.db.Exec("UPDATE timeslots SET is_booked = $2 WHERE id = $1", id, booked)
	return err
}

func (r *TimeslotRepository) GenerateSlots(
	scheduleID int,
	start, end time.Time,
	step time.Duration,
) error {
	const q = `
  INSERT INTO timeslots (schedule_id, start_time, end_time)
  VALUES ($1, $2, $3)
  ON CONFLICT DO NOTHING`
	for t := start; t.Add(step).Before(end) || t.Add(step).Equal(end); t = t.Add(step) {
		if _, err := r.db.Exec(q, scheduleID, t, t.Add(step)); err != nil {
			return err
		}
	}
	return nil
}
