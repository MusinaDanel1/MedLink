package http

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Room хранит два соединения: doctor и patient
type Room struct {
	doctor  *websocket.Conn
	patient *websocket.Conn
	mu      sync.Mutex
}

var (
	rooms    = make(map[string]*Room)
	roomsMu  sync.Mutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// SignalingHandler поднимает WS-соединение для обмена SDP/ICE
// URL: /ws?appointment_id=42&role=doctor|patient
func SignalingHandler(c *gin.Context) {
	apptID := c.Query("appointment_id")
	role := c.Query("role")
	if apptID == "" || (role != "doctor" && role != "patient") {
		c.JSON(400, gin.H{"error": "bad appointment_id or role"})
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// получаем или создаём комнату
	roomsMu.Lock()
	room, ok := rooms[apptID]
	if !ok {
		room = &Room{}
		rooms[apptID] = room
	}
	roomsMu.Unlock()

	// регистрируем в комнатe
	room.mu.Lock()
	if role == "doctor" {
		room.doctor = ws
	} else {
		room.patient = ws
	}
	room.mu.Unlock()

	// читаем из ws и форвардим в peer
	var peer **websocket.Conn
	if role == "doctor" {
		peer = &room.patient
	} else {
		peer = &room.doctor
	}

	for {
		mt, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		room.mu.Lock()
		if *peer != nil {
			(*peer).WriteMessage(mt, msg)
		}
		room.mu.Unlock()
	}
	ws.Close()
}
