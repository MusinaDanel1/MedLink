package postgres

import (
	"database/sql"
	"telemed/internal/domain"
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
		RETURNING id, appointment_id, sender, content, attachmemt_url, sent_at`

	var created domain.Message
	err := r.db.QueryRow(
		query,
		m.AppointmentID,
		m.Sender,
		m.Content,
		m.AttachmentURL,
	).Scan(
		&created.ID,
		&created.AppointmentID,
		&created.Sender,
		&created.Content,
		&created.AttachmentURL,
		&created.SentAt,
	)
	if err != nil {
		return domain.Message{}, err
	}
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
		if err := rows.Scan(
			&m.ID,
			&m.AppointmentID,
			&m.Sender,
			&m.Content,
			&m.AttachmentURL,
			&m.SentAt,
		); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}
