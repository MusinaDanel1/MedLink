package http

import (
	"strconv"
	"telemed/internal/domain"

	"github.com/gin-gonic/gin"

	msgUC "telemed/internal/usecase"
)

type MessageHandler struct{ svc *msgUC.MessageService }

func NewMessageHandler(s *msgUC.MessageService) *MessageHandler {
	return &MessageHandler{svc: s}
}

func (h *MessageHandler) List(c *gin.Context) {
	apptID, _ := strconv.Atoi(c.Param("id"))
	msgs, err := h.svc.List(apptID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, msgs)
}

func (h *MessageHandler) Create(c *gin.Context) {
	apptID, _ := strconv.Atoi(c.Param("id"))
	var m domain.Message
	if err := c.BindJSON(&m); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	m.AppointmentID = apptID
	saved, err := h.svc.Create(m)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, saved)
}
