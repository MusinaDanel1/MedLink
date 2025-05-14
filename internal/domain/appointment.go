package domain

import "time"

type Appointment struct {
	ID         int
	PatientID  int
	DoctorID   int
	ServiceID  int
	TimeSlotID int
	CreatedAt  time.Time
}

type TimeSlot struct {
	ID              int
	DoctorID        int
	AppointmentTime time.Time
	IsBooked        bool
}

type AppointmentRepository interface {
	CreateAppointment(a Appointment) error
	MarkTimeslotAsBooked(timeslotID int) error
}
