package postgres

import (
	"database/sql"
	"telemed/internal/domain"
)

type PatientRepository struct {
	db *sql.DB
}

func NewPatientRepository(db *sql.DB) *PatientRepository {
	return &PatientRepository{db: db}
}

func (r *PatientRepository) GetByTelegramID(telegramID int64) (*domain.Patient, error) {
	query := `SELECT id, full_name, iin, telegram_id FROM patients WHERE telegram_id = $1`
	row := r.db.QueryRow(query, telegramID)

	var p domain.Patient
	if err := row.Scan(&p.ID, &p.FullName, &p.IIN, &p.TelegramID); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PatientRepository) RegisterPatient(fullName, iin string, telegramID int64) error {
	_, err := r.db.Exec(`INSERT INTO patients (full_name, iin, telegram_id) VALUES ($1, $2, $3)`, fullName, iin, telegramID)
	return err
}

func (r *PatientRepository) GetAll() ([]domain.Patient, error) {
	query := `SELECT id, full_name, iin, telegram_id FROM patients ORDER BY id`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patients []domain.Patient
	for rows.Next() {
		var p domain.Patient
		if err := rows.Scan(&p.ID, &p.FullName, &p.IIN, &p.TelegramID); err != nil {
			return nil, err
		}
		patients = append(patients, p)
	}

	return patients, nil
}
func (pr *PatientRepository) GetIDByChatID(telegramID int64) (int, error) {
	var id int
	err := pr.db.QueryRow("SELECT id FROM patients WHERE telegram_id = $1", telegramID).Scan(&id)
	return id, err
}
