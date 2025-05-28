package domain

import "time"

type Appointment struct {
	ID         int       `json:"id"`
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
	AppointmentID int            `json:"appointmentId"`
	Complaints    string         `json:"complaints"`
	Diagnosis     string         `json:"diagnosis"`
	Assignment    string         `json:"assignment"`
	Prescriptions []Prescription `json:"prescriptions"`
}

type Prescription struct {
	Medication string
	Dosage     string
	Schedule   string
}

type NotificationData struct {
	AppointmentID int
	PatientChatID int64
	DoctorName    string
	ServiceName   string
	StartTime     time.Time
	Language      string
}

type CompleteAppointmentRequest struct {
	Complaints    string `json:"complaints"`
	Diagnosis     string `json:"diagnosis"`
	Assignment    string `json:"assignment"`
	Prescriptions []struct {
		Med      string `json:"med"`
		Dose     string `json:"dose"`
		Schedule string `json:"schedule"`
	} `json:"prescriptions"`
}

type AppointmentRepository interface {
	CreateAppointment(a Appointment) (int, error)
	MarkTimeslotAsBooked(timeslotID int) error
	GetAppointmentByID(id int) (*Appointment, error)
	GetPatientDetailsByID(patientID int) (map[string]interface{}, error)
	CompleteAppointment(details AppointmentDetails) error
	ListBySchedules(scheduleIDs []int) ([]Appointment, error)
	UpdateStatus(id int, status string) error
	FetchDetails(apptID int) (AppointmentDetails, error)
	GetUpcomingAppointments(from, to time.Time) ([]NotificationData, error)
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
	GetByID(id int) (*TimeSlot, error)
}
