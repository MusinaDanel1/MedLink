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
	openai      usecase.Service
}

func NewBotHandler(bot *tgbotapi.BotAPI, patient *usecase.PatientService, doctor *usecase.DoctorService, appointment *usecase.AppointmentService, openai usecase.Service) *BotHandler {
	return &BotHandler{
		bot:         bot,
		patient:     patient,
		doctor:      doctor,
		appointment: appointment,
		state:       make(map[int64]string),
		temp:        make(map[int64]map[string]string),
		openai:      openai,
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

	// Route all messages through HandleMessage
	h.HandleMessage(update.Message)
}

func (h *BotHandler) HandleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	// Handle /start command
	if text == "/start" {
		h.handleStart(chatID)
		return
	}

	// Если ожидание текста от пользователя для ChatGPT
	if h.state[chatID] == "ai_consultation_waiting" {
		reply, err := h.openai.AskChatGPT(text)
		if err != nil {
			h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при обращении к ИИ: "+err.Error()))
			return
		}

		// Отправляем ответ от ИИ с поддержкой Markdown
		aiMsg := tgbotapi.NewMessage(chatID, reply)
		aiMsg.ParseMode = "Markdown"
		h.bot.Send(aiMsg)
		return
	}

	// Handle regular keyboard buttons
	switch text {
	case "📅 Записаться к врачу":
		h.handleBookingStart(chatID)
	case "💬 Консультация с ИИ":
		h.state[chatID] = "ai_consultation_waiting"
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, опишите вашу жалобу, и я проконсультирую вас.")
		h.bot.Send(msg)
	default:
		// If it's not a command or button, handle as user input
		h.handleUserInput(chatID, text)
	}
}
