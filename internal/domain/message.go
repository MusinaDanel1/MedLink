package domain

import "time"

type Message struct {
	ID            int
	AppointmentID int
	Sender        string // "patient", "doctor", "bot"
	Content       string
	AttachmentURL string
	SentAt        time.Time
}

type MessageRepository interface {
	Create(m Message) (Message, error)
	ListByAppointment(appointmentID int) ([]Message, error)
}
