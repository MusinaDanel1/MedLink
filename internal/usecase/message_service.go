package usecase

import "telemed/internal/domain"

type MessageService struct {
	repo domain.MessageRepository
}

func NewMessageService(repo domain.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) List(appointmentID int) ([]domain.Message, error) {
	return s.repo.ListByAppointment(appointmentID)
}

func (s *MessageService) Create(m domain.Message) (domain.Message, error) {
	return s.repo.Create(m)
}

func (s *MessageService) AppointmentExists(appointmentID int) (bool, error) {
	return s.repo.AppointmentExists(appointmentID)
}
