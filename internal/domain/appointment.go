package domain

import "time"

type Appointment struct {
	ID         int       `json:"id"`
	DoctorID   int       `json:"doctorId"`
	ServiceID  int       `json:"serviceId"`
	TimeslotID int       `json:"timeslotId"`
	PatientID  int       `json:"patientId"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	VideoURL   string    `json:"video_url,omitempty"`
}

type Schedule struct {
	ID        int       `json:"id"`
	DoctorID  int       `json:"doctor_id"`
	ServiceID int       `json:"service_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Color     string    `json:"color"`
	Visible   bool      `json:"visible"`
}

type TimeSlot struct {
	ID         int
	ScheduleID int
	StartTime  time.Time
	EndTime    time.Time
	IsBooked   bool
	CreatedAt  time.Time
}

type AppointmentDetails struct {
	AppointmentID int
	Complaints    string
	Diagnosis     string
	Assignment    string
	Prescriptions []Prescription
}

type Prescription struct {
	Medication string
	Dosage     string
	Schedule   string
}

type AppointmentRepository interface {
	CreateAppointment(a Appointment) error
	MarkTimeslotAsBooked(timeslotID int) error
	GetAppointmentByID(id int) (*Appointment, error)
	GetPatientDetailsByID(patientID int) (map[string]interface{}, error)
	CompleteAppointment(details AppointmentDetails) error
	ListBySchedules(scheduleIDs []int) ([]Appointment, error)
	UpdateStatus(id int, status string) error
}

type ScheduleRepository interface {
	Create(s *Schedule) error
	GetByID(id int) (*Schedule, error)
	ListByDoctor(doctorID int) ([]Schedule, error)
	ToggleVisible(id int, visible bool) error
}

type TimeslotRepository interface {
	GetOrCreate(scheduleID int, start, end time.Time) (*TimeSlot, error)
	MarkBooked(id int, booked bool) error
	GenerateSlots(scheduleID int, start, end time.Time, step time.Duration) error
}
