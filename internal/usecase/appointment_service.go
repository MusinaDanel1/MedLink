package usecase

import (
	"errors"
	"fmt"
	"medlink/internal/domain"
	"time"
)

var ErrSlotBooked = errors.New("timeslot already booked")

type AppointmentService struct {
	SRepo    domain.ScheduleRepository
	TRepo    domain.TimeslotRepository
	repo     domain.AppointmentRepository
	videoSvc *VideoService
}

func NewAppointmentService(r domain.AppointmentRepository, sr domain.ScheduleRepository,
	tr domain.TimeslotRepository, vs *VideoService) *AppointmentService {
	return &AppointmentService{repo: r, SRepo: sr, TRepo: tr, videoSvc: vs}
}

func (u *AppointmentService) BookAppointment(
	scheduleID, patientID int,
	start, end time.Time,
) (int, error) {
	sch, err := u.SRepo.GetByID(scheduleID)
	if err != nil {
		return 0, err
	}
	ts, err := u.TRepo.GetOrCreate(scheduleID, start, end)
	if err != nil {
		return 0, err
	}
	if ts.IsBooked {
		return 0, ErrSlotBooked
	}
	ap := domain.Appointment{
		// DoctorID:   sch.DoctorID, // Removed
		ServiceID:  sch.ServiceID,
		TimeslotID: ts.ID,
		PatientID:  patientID,
		Status:     "Записан",
	}
	apptID, err := u.repo.CreateAppointment(ap)
	if err != nil {
		return 0, err
	}
	if err := u.TRepo.MarkBooked(ts.ID, true); err != nil {
		return 0, err
	}
	return apptID, nil
}

func (s AppointmentService) GetAppointmentByID(id int) (*domain.Appointment, error) {
	return s.repo.GetAppointmentByID(id)
}

func (s AppointmentService) GetPatientDetailsByID(patientID int) (map[string]interface{}, error) {
	return s.repo.GetPatientDetailsByID(patientID)
}

func (s *AppointmentService) CompleteAppointment(
	details domain.AppointmentDetails,
) error {
	return s.repo.CompleteAppointment(details)
}

func (u *AppointmentService) ListBySchedules(scheduleIDs []int) ([]domain.Appointment, error) {
	return u.repo.ListBySchedules(scheduleIDs)
}

func (u *AppointmentService) AcceptAppointment(apptID int) (string, error) {
	// 1) Меняем статус в appointments
	if err := u.repo.UpdateStatus(apptID, "Принят"); err != nil {
		return "", err
	}
	// 2) Создаём видеосессию
	vs, err := u.videoSvc.StartSession(apptID)
	if err != nil {
		return "", err
	}
	return vs.VideoURL, nil
}

func (s *AppointmentService) GetAppointmentDetails(id int) (domain.AppointmentDetails, error) {
	return s.repo.FetchDetails(id)
}

func (s *AppointmentService) EndCall(appointmentID int) error {
	return s.repo.UpdateStatus(appointmentID, "Завершен")
}

func (s *AppointmentService) GetAppointmentStatus(appointmentID int) (string, error) {
	appt, err := s.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return "", err
	}
	return appt.Status, nil
}

func (s *AppointmentService) GetUpcomingAppointments(
	from, to time.Time,
) ([]domain.NotificationData, error) {
	notifications, err := s.repo.GetUpcomingAppointments(from, to)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (s *AppointmentService) GetScheduleByTimeslotID(timeslotID int) (*domain.Schedule, error) {
	timeslot, err := s.TRepo.GetByID(timeslotID)
	if err != nil {
		// Consider wrapping the error for more context if that's a project pattern
		// e.g., return nil, fmt.Errorf("error fetching timeslot %d: %w", timeslotID, err)
		// For now, returning a more specific error if it's sql.ErrNoRows, otherwise the original error.
		// This assumes TRepo.GetByID returns sql.ErrNoRows for not found.
		// If TRepo.GetByID already maps to a domain error, this check might change.
		// Also, database/sql should be imported if checking sql.ErrNoRows directly.
		// For simplicity and consistency with other parts of the code that propagate errors directly:
		return nil, fmt.Errorf("error fetching timeslot %d: %w", timeslotID, err)
	}

	// The check `if timeslot == nil` is usually not needed if TRepo.GetByID returns an error (like sql.ErrNoRows) when not found.
	// If GetByID could return (nil, nil), then this check would be important.
	// Based on typical repository patterns, err != nil would cover "not found".

	if timeslot.ScheduleID == 0 { // Check if ScheduleID is valid
		return nil, fmt.Errorf("timeslot %d has an invalid ScheduleID (0)", timeslotID)
	}

	schedule, err := s.SRepo.GetByID(timeslot.ScheduleID)
	if err != nil {
		// Similar error handling considerations as above for SRepo.GetByID
		return nil, fmt.Errorf("error fetching schedule %d for timeslot %d: %w", timeslot.ScheduleID, timeslotID, err)
	}

	// Similar to timeslot, if SRepo.GetByID returns an error for "not found", this check might be redundant.
	// If it can return (nil, nil), then this is needed.
	if schedule == nil {
		return nil, fmt.Errorf("schedule %d not found for timeslot %d", timeslot.ScheduleID, timeslotID)
	}

	return schedule, nil
}
