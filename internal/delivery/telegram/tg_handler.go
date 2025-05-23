package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

		// Return to main menu after AI consultation
		delete(h.state, chatID)
		h.sendMainMenu(chatID)
		return
	}

	// Handle regular keyboard buttons
	switch text {
	case "📅 Записаться к врачу":
		// Start the booking flow
		h.handleBookingStart(chatID)
	case "💬 Консультация с ИИ":
		h.state[chatID] = "ai_consultation_waiting"
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, опишите вашу жалобу, и я проконсультирую вас.")
		h.bot.Send(msg)
	default:
		// If not a recognized command, show the main menu
		if h.state[chatID] == "" {
			h.sendMainMenu(chatID)
		} else {
			// Process as input for the current state
			h.handleUserInput(chatID, text)
		}
	}
}

func (h *BotHandler) SendReport(chatID int64, pdfBytes []byte, apptID int) error {
	doc := tgbotapi.FileBytes{
		Name:  fmt.Sprintf("report-%d.pdf", apptID),
		Bytes: pdfBytes,
	}
	// в v5 send raw bytes через NewDocument:
	msg := tgbotapi.NewDocument(chatID, doc)
	_, err := h.bot.Send(msg)
	return err
}

// internal/delivery/telegram/bot_handler.go
func (h *BotHandler) sendVideoLink(chatID int64, apptID int) {
	// 1) Формируем URL API
	apiURL := fmt.Sprintf("http://localhost:8080/api/appointments/%d/accept", apptID)
	// 2) Делаем PUT-запрос
	req, _ := http.NewRequest("PUT", apiURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Не удалось получить ссылку."))
		return
	}
	defer resp.Body.Close()
	var body struct {
		VideoUrl string `json:"videoUrl"`
	}
	json.NewDecoder(resp.Body).Decode(&body)

	link := body.VideoUrl
	u, err := url.Parse(link)
	if err != nil {
		// невалидный URL — просто добавим роль вручную
		if strings.Contains(link, "?") {
			link += "&role=patient"
		} else {
			link += "?role=patient"
		}
	} else {
		q := u.Query()
		q.Set("role", "patient")
		u.RawQuery = q.Encode()
		link = u.String()
	}

	h.bot.Send(tgbotapi.NewMessage(chatID,
		"Через 5 минут видеозвонок по ссылке:\n"+link))

}
