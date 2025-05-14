package postgres

import (
	"database/sql"
	"telemed/internal/domain"
)

type AppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) CreateAppointment(app domain.Appointment) error {
	_, err := r.db.Exec(
		"INSERT INTO appointments (patient_id, doctor_id, service_id, timeslots_id) VALUES ($1, $2, $3, $4)",
		app.PatientID, app.DoctorID, app.ServiceID, app.TimeSlotID,
	)
	return err
}

func (r *AppointmentRepository) MarkTimeslotAsBooked(timeslotID int) error {
	_, err := r.db.Exec(`UPDATE timeslots SET is_booked = true WHERE id = $1`, timeslotID)
	return err
}
