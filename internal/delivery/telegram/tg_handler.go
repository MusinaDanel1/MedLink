package telegram

import (
	"telemed/internal/usecase"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot         *tgbotapi.BotAPI
	patient     *usecase.PatientService
	doctor      *usecase.DoctorService
	appointment *usecase.AppointmentService
	state       map[int64]string
	temp        map[int64]map[string]string
}

func NewBotHandler(bot *tgbotapi.BotAPI, patient *usecase.PatientService, doctor *usecase.DoctorService, appointment *usecase.AppointmentService) *BotHandler {
	return &BotHandler{
		bot:         bot,
		patient:     patient,
		doctor:      doctor,
		appointment: appointment,
		state:       make(map[int64]string),
		temp:        make(map[int64]map[string]string),
	}
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		h.handleCallback(update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	msgText := update.Message.Text

	switch msgText {
	case "/start":
		h.handleStart(chatID)
	default:
		h.handleUserInput(chatID, msgText)
	}
}
