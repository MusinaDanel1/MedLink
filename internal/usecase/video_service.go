package usecase

import (
	"telemed/internal/domain"
)

type VideoService struct {
	repo domain.VideoSessionRepository
}

func NewVideoService(r domain.VideoSessionRepository) *VideoService {
	return &VideoService{repo: r}
}

// func (s *VideoService) StartSession(apptID int) (domain.VideoSession, error) {
// 	room := fmt.Sprintf("appointment-%d-%d", apptID, time.Now().Unix())
// 	vs := domain.VideoSession{AppointmentID: apptID, VideoURL: "https://meet.jit.si" + room, RoomName: room}
// }
