package postgres

import (
	"database/sql"
	"medlink/internal/domain"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(m domain.Message) (domain.Message, error) {
	const query = `
	INSERT INTO messages (
	    appointment_id, sender, content, attachment_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, appointment_id, sender, content, attachment_url, sent_at`

	var created domain.Message
	var attachmentURL sql.NullString
	err := r.db.QueryRow(
		query,
		m.AppointmentID,
		m.Sender,
		m.Content,
		sql.NullString{String: m.AttachmentURL, Valid: m.AttachmentURL != ""},
	).Scan(
		&created.ID,
		&created.AppointmentID,
		&created.Sender,
		&created.Content,
		&attachmentURL,
		&created.SentAt,
	)
	if err != nil {
		return domain.Message{}, err
	}
	created.AttachmentURL = attachmentURL.String
	return created, nil
}

func (r *MessageRepository) ListByAppointment(appointmentID int) ([]domain.Message, error) {
	const query = `
	SELECT id, appointment_id, sender, content, attachment_url, sent_at
	FROM messages
	WHERE appointment_id = $1
	ORDER BY sent_at ASC`

	rows, err := r.db.Query(query, appointmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.Message
	for rows.Next() {
		var m domain.Message
		var attachmentURL sql.NullString
		if err := rows.Scan(
			&m.ID,
			&m.AppointmentID,
			&m.Sender,
			&m.Content,
			&attachmentURL,
			&m.SentAt,
		); err != nil {
			return nil, err
		}
		m.AttachmentURL = attachmentURL.String
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (r *MessageRepository) AppointmentExists(appointmentID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM appointments WHERE id = $1)`
	err := r.db.QueryRow(query, appointmentID).Scan(&exists)
	return exists, err
}
