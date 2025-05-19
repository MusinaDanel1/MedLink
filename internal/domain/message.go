package domain

import "time"

type Message struct {
	ID            int       `json:"id"`
	AppointmentID int       `json:"appointment_id"`
	Sender        string    `json:"sender"` // "patient", "doctor", "bot"
	Content       string    `json:"content"`
	AttachmentURL string    `json:"attachment_url"`
	SentAt        time.Time `json:"sent_at"`
}

type MessageRepository interface {
	Create(m Message) (Message, error)
	ListByAppointment(appointmentID int) ([]Message, error)
	AppointmentExists(appointmentID int) (bool, error)
}
