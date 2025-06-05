package postgres

import (
	"database/sql"
	"medlink/internal/domain"
)

type VideoSessionRepository struct {
	db *sql.DB
}

func NewVideoSessionRepository(db *sql.DB) *VideoSessionRepository {
	return &VideoSessionRepository{db: db}
}

func (r *VideoSessionRepository) Create(vs domain.VideoSession) (domain.VideoSession, error) {
	const query = `
	INSERT INTO video_sessions (appointment_id, room_name, video_url)
	VALUES ($1, $2, $3)
	RETURNING id, appointment_id, room_name, video_url, started_at, ended_at`

	var created domain.VideoSession
	err := r.db.QueryRow(
		query,
		vs.AppointmentID,
		vs.RoomName,
		vs.VideoURL,
	).Scan(
		&created.ID,
		&created.AppointmentID,
		&created.RoomName,
		&created.VideoURL,
		&created.StartedAt,
		&created.EndedAd,
	)
	if err != nil {
		return domain.VideoSession{}, err
	}
	return created, nil
}

func (r *VideoSessionRepository) End(vsID int) error {
	const query = `
	UPDATE video_sessions
	SET ended_at = NOW()
	WHERE id = $1`

	_, err := r.db.Exec(query, vsID)
	return err
}

func (r *VideoSessionRepository) FindByIDAppointment(apptID int) (domain.VideoSession, error) {
	const query = `
	  SELECT id, appointment_id, room_name, video_url, started_at, ended_at
	  FROM video_sessions
	  WHERE appointment_id = $1
	  ORDER BY started_at DESC
	  LIMIT 1`

	var vs domain.VideoSession
	err := r.db.QueryRow(query, apptID).Scan(
		&vs.ID,
		&vs.AppointmentID,
		&vs.RoomName,
		&vs.VideoURL,
		&vs.StartedAt,
		&vs.EndedAd,
	)
	if err != nil {
		return domain.VideoSession{}, err
	}
	return vs, nil
}
