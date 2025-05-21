package usecase

import (
	"fmt"
	"telemed/internal/domain"
	"time"
)

type VideoService struct {
	repo domain.VideoSessionRepository
}

func NewVideoService(r domain.VideoSessionRepository) *VideoService {
	return &VideoService{repo: r}
}

func (s *VideoService) StartSession(apptID int) (domain.VideoSession, error) {
	// формируем уникальное имя комнаты
	roomName := fmt.Sprintf("appointment-%d-%d", apptID, time.Now().Unix())
	videoURL := "https://meet.jit.si/" + roomName

	vs, err := s.repo.Create(domain.VideoSession{
		AppointmentID: apptID,
		RoomName:      roomName,
		VideoURL:      videoURL,
	})
	return vs, err
}
