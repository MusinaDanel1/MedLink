package usecase

import (
	"medlink/internal/domain"
)

// ScheduleService описывает бизнес-логику schedule
type ScheduleService interface {
	ListByDoctor(doctorID int) ([]domain.Schedule, error)
	Create(s *domain.Schedule) error
	ToggleVisible(id int, visible bool) error
}

type scheduleService struct {
	repo domain.ScheduleRepository
}

func NewScheduleService(r domain.ScheduleRepository) ScheduleService {
	return &scheduleService{repo: r}
}

func (s *scheduleService) ListByDoctor(doctorID int) ([]domain.Schedule, error) {
	return s.repo.ListByDoctor(doctorID)
}

func (s *scheduleService) Create(sch *domain.Schedule) error {
	return s.repo.Create(sch)
}

func (s *scheduleService) ToggleVisible(id int, visible bool) error {
	return s.repo.ToggleVisible(id, visible)
}
