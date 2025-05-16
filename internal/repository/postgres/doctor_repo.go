package postgres

import (
	"database/sql"
	"telemed/internal/domain"
)

type DoctorRepository struct {
	db *sql.DB
}

func NewDoctorRepository(db *sql.DB) *DoctorRepository {
	return &DoctorRepository{db: db}
}

func (r *DoctorRepository) CreateDoctor(fullName string, specializationID int) error {
	_, err := r.db.Exec(
		"INSERT INTO doctors (full_name, specialization_id) VALUES ($1, $2)",
		fullName, specializationID,
	)
	return err
}

func (r *DoctorRepository) GetAllSpecializations() ([]domain.Specialization, error) {
	rows, err := r.db.Query("SELECT id,name FROM specializations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var specs []domain.Specialization
	for rows.Next() {
		var s domain.Specialization
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}
	return specs, nil
}

func (r *DoctorRepository) GetDoctorsBySpecialization(specializationID int) ([]domain.Doctor, error) {
	rows, err := r.db.Query(`SELECT id, full_name, specialization_id FROM doctors WHERE specialization_id = $1`, specializationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctors []domain.Doctor
	for rows.Next() {
		var d domain.Doctor
		if err := rows.Scan(&d.ID, &d.FullName, &d.SpecializationID); err != nil {
			return nil, err
		}
		doctors = append(doctors, d)
	}
	return doctors, nil
}

func (r *DoctorRepository) GetServicesByDoctor(doctorID int) ([]domain.Service, error) {
	rows, err := r.db.Query("SELECT id, doctor_id, name FROM services WHERE doctor_id = $1", doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []domain.Service
	for rows.Next() {
		var s domain.Service
		s.DoctorID = doctorID
		if err := rows.Scan(&s.ID, &s.DoctorID, &s.Name); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func (r *DoctorRepository) GetAvailableTimeSlots(doctorID int) ([]domain.TimeSlot, error) {
	rows, err := r.db.Query(`SELECT id, doctor_id, appointment_time, is_booked FROM timeslots WHERE doctor_id = $1 AND is_booked = false`, doctorID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var timeslots []domain.TimeSlot
	for rows.Next() {
		var t domain.TimeSlot
		if err := rows.Scan(&t.ID, &t.DoctorID, &t.AppointmentTime, &t.IsBooked); err != nil {
			return nil, err
		}
		timeslots = append(timeslots, t)
	}

	return timeslots, nil
}
