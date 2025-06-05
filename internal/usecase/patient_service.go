package usecase

import (
	"medlink/internal/domain"
)

type PatientService struct {
	repo domain.PatientRepository
}

func NewPatientService(r domain.PatientRepository) *PatientService {
	return &PatientService{repo: r}
}

func (s *PatientService) FindOrRegister(chatID int64, fullName, iin string) error {
	return s.repo.RegisterPatient(fullName, iin, chatID)
}

func (s *PatientService) Exists(chatID int64) bool {
	patient, err := s.repo.GetByTelegramID(chatID)
	return err == nil && patient != nil
}

func (s *PatientService) GetIDByChatID(chatID int64) (int, error) {
	patient, err := s.repo.GetByTelegramID(chatID)
	if err != nil {
		return 0, err
	}
	return patient.ID, nil
}

func (s *PatientService) GetAll() ([]domain.Patient, error) {
	return s.repo.GetAll()
}

func (ps *PatientService) GetByTelegramID(telegramID int64) (*domain.Patient, error) {
	return ps.repo.GetByTelegramID(telegramID)
}
