package usecase

import (
	"telemed/internal/domain"
)

type PatientService struct {
	repo domain.PatientRepository
}

func NewPatientService(r domain.PatientRepository) *PatientService {
	return &PatientService{repo: r}
}

func (s *PatientService) FindOrRegister(telegramID int64, fullName, iin string) error {
	_, err := s.repo.GetByTelegramID(telegramID)
	if err == nil {
		return nil
	}
	return s.repo.RegisterPatient(fullName, iin, telegramID)
}

func (s *PatientService) Exists(telegramID int64) bool {
	_, err := s.repo.GetByTelegramID(telegramID)
	return err == nil
}
