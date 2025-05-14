package usecase

import "telemed/internal/domain"

type AppointmentService struct {
	repo domain.AppointmentRepository
}

func NewAppointmentService(r domain.AppointmentRepository) *AppointmentService {
	return &AppointmentService{repo: r}
}

func (s AppointmentService) BookAppointment(patientID, doctorID, serviceID, timeslotID int) error {
	err := s.repo.CreateAppointment(domain.Appointment{
		PatientID:  patientID,
		DoctorID:   doctorID,
		ServiceID:  serviceID,
		TimeSlotID: timeslotID,
	})

	if err != nil {
		return err
	}

	return s.repo.MarkTimeslotAsBooked(timeslotID)
}
