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
	if vs, err := s.repo.FindByIDAppointment(apptID); err == nil {
		return vs, nil
	}
	roomName := fmt.Sprintf("appointment-%d-%d",
		apptID, time.Now().Unix())

	// 3) собираем относительный путь на ваш WebRTC-рум
	//    фронт будет его превращать в абсолютный URL
	videoURL := fmt.Sprintf("/webrtc/room.html?appointment_id=%d",
		apptID)

	vs, err := s.repo.Create(domain.VideoSession{
		AppointmentID: apptID,
		RoomName:      roomName,
		VideoURL:      videoURL,
	})
	return vs, err
}
