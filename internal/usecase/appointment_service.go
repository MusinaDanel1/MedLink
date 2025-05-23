package usecase

import (
	"errors"
	"telemed/internal/domain"
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
		DoctorID:   sch.DoctorID,
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
