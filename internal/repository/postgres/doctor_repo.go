package postgres

import (
	"database/sql"
	"fmt"
	"log"
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
	// First try to get doctor-specific services
	log.Printf("[DB_DEBUG] Getting services for doctor ID: %d", doctorID)
	query := "SELECT id, doctor_id, name FROM services WHERE doctor_id = $1"
	log.Printf("[DB_DEBUG] Running query: %s with doctorID=%d", query, doctorID)

	rows, err := r.db.Query(query, doctorID)
	if err != nil {
		log.Printf("[DB_DEBUG] Error querying services: %v", err)
		return nil, err
	}
	defer rows.Close()

	var services []domain.Service
	for rows.Next() {
		var s domain.Service
		s.DoctorID = doctorID
		if err := rows.Scan(&s.ID, &s.DoctorID, &s.Name); err != nil {
			log.Printf("[DB_DEBUG] Error scanning service row: %v", err)
			return nil, err
		}
		services = append(services, s)
		log.Printf("[DB_DEBUG] Found service: ID=%d, Name=%s, DoctorID=%d", s.ID, s.Name, s.DoctorID)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[DB_DEBUG] Error after reading rows: %v", err)
		return nil, err
	}

	// If no doctor-specific services were found, try to get some general services
	if len(services) == 0 {
		log.Printf("[DB_DEBUG] No services found for doctor ID %d, looking for general services", doctorID)
		// Look for services that have null or 0 doctor_id (general services)
		generalQuery := "SELECT id, COALESCE(doctor_id, 0), name FROM services WHERE doctor_id IS NULL OR doctor_id = 0"
		log.Printf("[DB_DEBUG] Running general query: %s", generalQuery)

		rows, err := r.db.Query(generalQuery)
		if err != nil {
			log.Printf("[DB_DEBUG] Error querying general services: %v", err)
			return services, nil // Return empty list rather than error
		}
		defer rows.Close()

		for rows.Next() {
			var s domain.Service
			var nullableDoctorID sql.NullInt64
			if err := rows.Scan(&s.ID, &nullableDoctorID, &s.Name); err != nil {
				log.Printf("[DB_DEBUG] Error scanning general service row: %v", err)
				continue // Skip invalid records
			}
			// Assign the current doctor ID to this general service
			s.DoctorID = doctorID
			services = append(services, s)
			log.Printf("[DB_DEBUG] Found general service: ID=%d, Name=%s, reassigned to DoctorID=%d", s.ID, s.Name, s.DoctorID)
		}

		if err = rows.Err(); err != nil {
			log.Printf("[DB_DEBUG] Error after reading general rows: %v", err)
		}

		// If we still have no services, that's fine - we'll add defaults at the handler level
		log.Printf("[DB_DEBUG] After checking general services, found a total of %d services", len(services))
	} else {
		log.Printf("[DB_DEBUG] Found %d doctor-specific services", len(services))
	}

	return services, nil
}

func (r *DoctorRepository) GetAvailableTimeSlots(doctorID int) ([]domain.TimeSlot, error) {
	query := `
		SELECT t.id, t.schedule_id, t.start_time, t.end_time, t.is_booked, t.created_at
		FROM timeslots t
		JOIN schedules s ON t.schedule_id = s.id
		WHERE s.doctor_id = $1 AND t.is_booked = false
	`
	rows, err := r.db.Query(query, doctorID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var timeslots []domain.TimeSlot
	for rows.Next() {
		var t domain.TimeSlot
		if err := rows.Scan(&t.ID, &t.ScheduleID, &t.StartTime, &t.EndTime, &t.IsBooked, &t.CreatedAt); err != nil {
			return nil, err
		}
		timeslots = append(timeslots, t)
	}

	return timeslots, nil
}

func (r *DoctorRepository) GetAllDiagnoses() ([]domain.Diagnosis, error) {
	query := `SELECT id, code, name, description FROM diagnoses ORDER BY code`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diagnoses []domain.Diagnosis
	for rows.Next() {
		var d domain.Diagnosis
		if err := rows.Scan(&d.ID, &d.Code, &d.Name, &d.Description); err != nil {
			return nil, err
		}
		diagnoses = append(diagnoses, d)
	}

	return diagnoses, nil
}

func (r *DoctorRepository) GetDoctorByID(id int) (*domain.Doctor, error) {
	query := `SELECT id, full_name, specialization_id FROM doctors WHERE id = $1`

	row := r.db.QueryRow(query, id)

	var doctor domain.Doctor
	err := row.Scan(&doctor.ID, &doctor.FullName, &doctor.SpecializationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("doctor with ID %d not found", id)
		}
		return nil, err
	}

	return &doctor, nil
}

func (r *DoctorRepository) GetAllServices() ([]domain.Service, error) {
	rows, err := r.db.Query("SELECT id, doctor_id, name FROM services")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []domain.Service
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(&s.ID, &s.DoctorID, &s.Name); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func (r *DoctorRepository) GetDoctorByIIN(iin string) (*domain.Doctor, error) {
	query := `
		SELECT d.id, d.full_name, d.specialization_id 
		FROM doctors d
		JOIN users u ON d.full_name = u.full_name
		WHERE u.iin = $1
	`

	row := r.db.QueryRow(query, iin)

	var doctor domain.Doctor
	err := row.Scan(&doctor.ID, &doctor.FullName, &doctor.SpecializationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("doctor with IIN %s not found", iin)
		}
		return nil, err
	}

	return &doctor, nil
}

func (r *DoctorRepository) CreateService(doctorID int, name string) (*domain.Service, error) {
	// First check if the doctor exists
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM doctors WHERE id = $1)", doctorID).Scan(&exists)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("doctor with ID %d doesn't exist", doctorID)
	}

	// Check if a service with this name already exists for this doctor
	var existingID int
	err = r.db.QueryRow("SELECT id FROM services WHERE doctor_id = $1 AND name = $2", doctorID, name).Scan(&existingID)
	if err == nil {
		// Service already exists, return it
		return &domain.Service{
			ID:       existingID,
			Name:     name,
			DoctorID: doctorID,
		}, nil
	} else if err != sql.ErrNoRows {
		// Some other error occurred
		return nil, err
	}

	// Service doesn't exist, create it
	var serviceID int
	err = r.db.QueryRow(
		"INSERT INTO services (doctor_id, name) VALUES ($1, $2) RETURNING id",
		doctorID, name,
	).Scan(&serviceID)

	if err != nil {
		return nil, err
	}

	return &domain.Service{
		ID:       serviceID,
		Name:     name,
		DoctorID: doctorID,
	}, nil
}

func (r *DoctorRepository) GetAllDoctors() ([]domain.Doctor, error) {
	query := "SELECT id, full_name, specialization_id FROM doctors ORDER BY id"

	rows, err := r.db.Query(query)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return doctors, nil
}

func (r *DoctorRepository) GetSpecializationName(id int) (string, error) {
	var name string
	err := r.db.QueryRow(
		`SELECT name FROM specializations WHERE id = $1`, id,
	).Scan(&name)
	return name, err
}
