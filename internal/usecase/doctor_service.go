package usecase

import "telemed/internal/domain"

type DoctorService struct {
	repo domain.DoctorRepository
}

func NewDoctorService(r domain.DoctorRepository) *DoctorService {
	return &DoctorService{repo: r}
}

func (s *DoctorService) GetAllSpecializations() ([]domain.Specialization, error) {
	return s.repo.GetAllSpecializations()
}

func (s *DoctorService) GetDoctorsBySpecialization(specializationID int) ([]domain.Doctor, error) {
	return s.repo.GetDoctorsBySpecialization(specializationID)
}

func (s *DoctorService) GetServicesByDoctor(doctorID int) ([]domain.Service, error) {
	return s.repo.GetServicesByDoctor(doctorID)
}

func (s *DoctorService) GetAvailableTimeSlots(doctorID int) ([]domain.TimeSlot, error) {
	return s.repo.GetAvailableTimeSlots(doctorID)
}
