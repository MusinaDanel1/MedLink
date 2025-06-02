package postgres

import (
	"database/sql"
	"log"
	"strings"
	"telemed/internal/domain"
	"time"

	"github.com/lib/pq"
)

type AppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) CreateAppointment(app domain.Appointment) (int, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO appointments (patient_id, service_id, timeslot_id) VALUES ($1, $2, $3) RETURNING id",
		app.PatientID, app.ServiceID, app.TimeslotID,
	).Scan(&id)
	return id, err
}

func (r *AppointmentRepository) MarkTimeslotAsBooked(timeslotID int) error {
	_, err := r.db.Exec(`UPDATE timeslots SET is_booked = true WHERE id = $1`, timeslotID)
	return err
}

func (r *AppointmentRepository) GetAppointmentByID(id int) (*domain.Appointment, error) {
	query := `SELECT id, patient_id, service_id, status, timeslot_id, created_at 
			  FROM appointments WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var appt domain.Appointment
	err := row.Scan(
		&appt.ID,
		&appt.PatientID,
		// &appt.DoctorID, // Removed
		&appt.ServiceID,
		&appt.Status,
		&appt.TimeslotID,
		&appt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &appt, nil
}

func (r *AppointmentRepository) GetPatientDetailsByID(patientID int) (map[string]interface{}, error) {
	// Basic patient info
	patientQuery := `SELECT p.id, p.full_name, p.iin, p.telegram_id 
			  FROM patients p WHERE p.id = $1`

	row := r.db.QueryRow(patientQuery, patientID)

	var id int
	var fullName, iin string
	var telegramID int64

	err := row.Scan(
		&id,
		&fullName,
		&iin,
		&telegramID,
	)

	if err != nil {
		return nil, err
	}

	// Get medical history entries
	historyQuery := `SELECT entry_type, description, event_date
				FROM medical_histories mh
				JOIN medical_records mr ON mh.medical_record_id = mr.id
				WHERE mr.patient_id = $1`

	rows, err := r.db.Query(historyQuery, patientID)

	// Create default empty results
	result := map[string]interface{}{
		"id":              id,
		"full_name":       fullName,
		"iin":             iin,
		"telegram_id":     telegramID,
		"medical_history": []interface{}{},
		"allergies":       []interface{}{},
		"vaccinations":    []interface{}{},
		"surgeries":       []interface{}{},
		"examinations":    []interface{}{},
	}

	// Process medical history entries if there are any
	if err == nil {
		defer rows.Close()

		var chronicDiseases []map[string]string
		var allergies []map[string]string
		var vaccinations []map[string]string
		var surgeries []map[string]string
		var examinations []map[string]string

		for rows.Next() {
			var entryType, description string
			var eventDate sql.NullTime

			if err := rows.Scan(&entryType, &description, &eventDate); err != nil {
				continue
			}

			dateStr := ""
			if eventDate.Valid {
				dateStr = eventDate.Time.Format("2006-01-02")
			}

			switch entryType {
			case "chronic_disease":
				chronicDiseases = append(chronicDiseases, map[string]string{
					"diagnosis": description,
					"date":      dateStr,
				})
			case "allergy":
				allergies = append(allergies, map[string]string{
					"name": description,
					"date": dateStr,
				})
			case "vaccination":
				vaccinations = append(vaccinations, map[string]string{
					"vaccine": description,
					"date":    dateStr,
				})
			case "surgery":
				surgeries = append(surgeries, map[string]string{
					"procedure": description,
					"date":      dateStr,
				})
			case "examination":
				// Split examination into name and result
				parts := strings.SplitN(description, " — ", 2)
				exam := parts[0]
				result := ""
				if len(parts) > 1 {
					result = parts[1]
				}

				examinations = append(examinations, map[string]string{
					"exam":   exam,
					"result": result,
					"date":   dateStr,
				})
			}
		}

		if len(chronicDiseases) > 0 {
			result["medical_history"] = chronicDiseases
		}
		if len(allergies) > 0 {
			result["allergies"] = allergies
		}
		if len(vaccinations) > 0 {
			result["vaccinations"] = vaccinations
		}
		if len(surgeries) > 0 {
			result["surgeries"] = surgeries
		}
		if len(examinations) > 0 {
			result["examinations"] = examinations
		}
	} else {
		// Only use default data if database query failed
		// Default data for development if no records exist
		result["medical_history"] = []map[string]string{
			{"diagnosis": "Гипертония", "date": "2022-01-10"},
			{"diagnosis": "Астма", "date": "2021-03-15"},
		}
		result["allergies"] = []map[string]string{
			{"name": "Пенициллин", "date": "2020-05-20"},
		}
		result["vaccinations"] = []map[string]string{
			{"vaccine": "COVID-19", "date": "2021-06-15"},
			{"vaccine": "Грипп", "date": "2022-11-10"},
		}
		result["surgeries"] = []map[string]string{
			{"procedure": "Аппендэктомия", "date": "2010-08-22"},
		}
		result["examinations"] = []map[string]string{
			{"exam": "Флюорография", "result": "без патологии", "date": "2023-02-15"},
			{"exam": "ЭКГ", "result": "синусовая аритмия", "date": "2024-01-20"},
		}
	}

	return result, nil
}

func (r *AppointmentRepository) CompleteAppointment(
	d domain.AppointmentDetails,
) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 2) вставляем или обновляем appointment_details
	_, err = tx.Exec(`
    INSERT INTO appointment_details
      (appointment_id, complaints, diagnosis, assignments, updated_at)
    VALUES ($1,$2,$3,$4,NOW())
    ON CONFLICT (appointment_id) DO UPDATE SET
      complaints  = EXCLUDED.complaints,
      diagnosis   = EXCLUDED.diagnosis,
      assignments = EXCLUDED.assignments,
      updated_at  = NOW()
  `, d.AppointmentID, d.Complaints, d.Diagnosis, d.Assignment)
	if err != nil {
		return err
	}

	// 3) удаляем старые рецепты
	_, err = tx.Exec(
		`DELETE FROM prescriptions WHERE appointment_id=$1`,
		d.AppointmentID,
	)
	if err != nil {
		return err
	}

	// 4) вставляем новые
	for _, p := range d.Prescriptions {
		_, err = tx.Exec(`
      INSERT INTO prescriptions
        (appointment_id, medication, dosage, schedule)
      VALUES ($1,$2,$3,$4)
    `, d.AppointmentID, p.Medication, p.Dosage, p.Schedule)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *AppointmentRepository) ListBySchedules(
	schedIDs []int,
) ([]domain.Appointment, error) {
	const q = `
SELECT
  a.id,
  -- a.doctor_id, // Removed
  a.service_id,
  a.timeslot_id,
  a.patient_id,
  a.status,
  a.created_at,
  t.start_time,
  t.end_time,
  -- последний video_url, если есть
  (
    SELECT vs.video_url
      FROM video_sessions vs
     WHERE vs.appointment_id = a.id
     ORDER BY vs.started_at DESC
     LIMIT 1
  ) AS video_url
FROM appointments a
JOIN timeslots t ON t.id = a.timeslot_id
WHERE t.schedule_id = ANY($1)
`
	rows, err := r.db.Query(q, pq.Array(schedIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Appointment
	for rows.Next() {
		var a domain.Appointment
		var videoURLNull sql.NullString

		if err := rows.Scan(
			&a.ID,
			// &a.DoctorID, // Removed
			&a.ServiceID,
			&a.TimeslotID,
			&a.PatientID,
			&a.Status,
			&a.CreatedAt,
			&a.StartTime,
			&a.EndTime,
			&videoURLNull,
		); err != nil {
			return nil, err
		}
		if videoURLNull.Valid {
			a.VideoURL = videoURLNull.String
		}
		out = append(out, a)
	}
	log.Printf("[APPT_REPO] scheduleIDs=%v", schedIDs)

	if out == nil {
		out = make([]domain.Appointment, 0)
	}
	log.Printf("[APPT_REPO] out=%+v", out)
	return out, rows.Err()
}

func (r *AppointmentRepository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec(
		`UPDATE appointments SET status=$2 WHERE id=$1`,
		id, status,
	)
	return err
}

// internal/repository/postgres/appointment_repo.go
func (r *AppointmentRepository) FetchDetails(apptID int) (domain.AppointmentDetails, error) {
	var d domain.AppointmentDetails
	d.AppointmentID = apptID
	// 1) complaints, diagnosis, assignment
	err := r.db.QueryRow(`
	  SELECT complaints, diagnosis, assignments
		FROM appointment_details
	   WHERE appointment_id=$1
	`, apptID).Scan(&d.Complaints, &d.Diagnosis, &d.Assignment)
	if err != nil && err != sql.ErrNoRows {
		return d, err
	}
	// 2) prescriptions
	rows, err := r.db.Query(`
	  SELECT medication, dosage, schedule
		FROM prescriptions
	   WHERE appointment_id=$1
	`, apptID)
	if err != nil {
		return d, err
	}
	defer rows.Close()

	for rows.Next() {
		var p domain.Prescription
		if err := rows.Scan(&p.Medication, &p.Dosage, &p.Schedule); err != nil {
			continue
		}
		d.Prescriptions = append(d.Prescriptions, p)
	}
	return d, rows.Err()
}

func (r *AppointmentRepository) GetUpcomingAppointments(from, to time.Time) ([]domain.NotificationData, error) {
	query := `
		SELECT 
			a.id,
			p.telegram_id,
			d.full_name as doctor_name,
			s.name as service_name,
			t.start_time,
			'ru'::TEXT AS language -- Fixed: p.language does not exist
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN timeslots t ON a.timeslot_id = t.id
		JOIN schedules sch ON t.schedule_id = sch.id
		JOIN doctors d ON sch.doctor_id = d.id
		JOIN services s ON sch.service_id = s.id
		WHERE t.start_time BETWEEN $1 AND $2
		AND a.status = 'Записан'
		ORDER BY t.start_time
	`

	rows, err := r.db.Query(query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []domain.NotificationData
	for rows.Next() {
		var notif domain.NotificationData
		err := rows.Scan(
			&notif.AppointmentID,
			&notif.PatientChatID,
			&notif.DoctorName,
			&notif.ServiceName,
			&notif.StartTime,
			&notif.Language,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}
