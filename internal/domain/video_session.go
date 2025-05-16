package domain

import "time"

type VideoSession struct {
	ID            int
	AppointmentID int
	RoomName      string
	VideoURL      string
	StartedAt     time.Time
	EndedAd       *time.Time
}

type VideoSessionRepository interface {
	Create(vs VideoSession) (VideoSession, error)
	End(vsID int) error
	FindByIDAppointment(apptID int) (VideoSession, error)
}
