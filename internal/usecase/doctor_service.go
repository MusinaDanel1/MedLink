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

func (s *DoctorService) CreateDoctor(fullName string, specializationID int) error {
	return s.repo.CreateDoctor(fullName, specializationID)
}

func (s *DoctorService) GetAllDiagnoses() ([]domain.Diagnosis, error) {
	return s.repo.GetAllDiagnoses()
}

func (s *DoctorService) GetAllServices() ([]domain.Service, error) {
	return s.repo.GetAllServices()
}

func (s *DoctorService) GetDoctorByID(id int) (*domain.Doctor, error) {
	return s.repo.GetDoctorByID(id)
}

func (s *DoctorService) GetDoctorByIIN(iin string) (*domain.Doctor, error) {
	return s.repo.GetDoctorByIIN(iin)
}

func (s *DoctorService) CreateService(doctorID int, name string) (*domain.Service, error) {
	return s.repo.CreateService(doctorID, name)
}

func (s *DoctorService) GetAllDoctors() ([]domain.Doctor, error) {
	return s.repo.GetAllDoctors()
}

func (s *DoctorService) GetSpecializationName(id int) (string, error) {
	return s.repo.GetSpecializationName(id)
}
