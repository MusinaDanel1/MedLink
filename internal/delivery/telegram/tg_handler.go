package telegram

import (
	"fmt"
	"os"
	"telemed/internal/usecase"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot                 *tgbotapi.BotAPI
	patientService      *usecase.PatientService
	doctorService       *usecase.DoctorService
	appointmentService  *usecase.AppointmentService
	notificationService *usecase.NotificationService
	state               map[int64]string
	temp                map[int64]map[string]string
	openai              OpenAIService
	userLang            map[int64]Language
	loc                 *Localization
}

type OpenAIService interface {
	AskChatGPT(text string) (string, error)
}

func NewBotHandler(
	bot *tgbotapi.BotAPI,
	patientService *usecase.PatientService,
	doctorService *usecase.DoctorService,
	appointmentService *usecase.AppointmentService,
	notificationService *usecase.NotificationService,
	openai OpenAIService,
) *BotHandler {
	handler := &BotHandler{
		bot:                 bot,
		patientService:      patientService,
		doctorService:       doctorService,
		appointmentService:  appointmentService,
		notificationService: notificationService,
		state:               make(map[int64]string),
		temp:                make(map[int64]map[string]string),
		openai:              openai,
		userLang:            make(map[int64]Language),
		loc:                 NewLocalization(),
	}

	return handler
}

func (h *BotHandler) SetNotificationService(ns *usecase.NotificationService) {
	h.notificationService = ns
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

	lang := h.getUserLanguage(chatID)
	if h.state[chatID] == "choosing_language" {
		h.handleLanguageSelection(chatID, text)
		return
	}

	// Если ожидание текста от пользователя для ChatGPT
	if h.state[chatID] == "ai_consultation_waiting" {
		reply, err := h.openai.AskChatGPT(text)
		if err != nil {
			h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "ai_error")+" "+err.Error()))
			return
		}

		aiMsg := tgbotapi.NewMessage(chatID, reply)
		aiMsg.ParseMode = "Markdown"
		h.bot.Send(aiMsg)

		delete(h.state, chatID)
		h.sendMainMenu(chatID)
		return
	}

	// Handle regular keyboard buttons
	switch text {
	case h.loc.Get(lang, "book_appointment"):
		h.handleBookingStart(chatID)
	case h.loc.Get(lang, "ai_consultation"):
		h.state[chatID] = "ai_consultation_waiting"
		msg := tgbotapi.NewMessage(chatID, h.loc.Get(lang, "ai_consultation_prompt"))
		h.bot.Send(msg)
	default:
		if h.state[chatID] == "" {
			h.sendMainMenu(chatID)
		} else {
			h.handleUserInput(chatID, text)
		}
	}
}

func (h *BotHandler) getUserLanguage(chatID int64) Language {
	if lang, ok := h.userLang[chatID]; ok {
		return lang
	}
	return LangRussian // Default to Russian
}

func (h *BotHandler) handleLanguageSelection(chatID int64, text string) {
	switch text {
	case h.loc.Get(LangRussian, "russian"):
		h.userLang[chatID] = LangRussian
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(LangRussian, "language_selected")))
	case h.loc.Get(LangKazakh, "kazakh"):
		h.userLang[chatID] = LangKazakh
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(LangKazakh, "language_selected")))
	default:
		// Invalid selection, ask again
		h.sendLanguageSelection(chatID)
		return
	}

	delete(h.state, chatID)

	// Check if user is registered
	isRegistered := h.patientService.Exists(chatID)
	if isRegistered {
		h.sendMainMenu(chatID)
	} else {
		h.startRegistration(chatID)
	}
}

func (h *BotHandler) sendLanguageSelection(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, h.loc.Get(LangRussian, "choose_language"))

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(h.loc.Get(LangRussian, "russian")),
			tgbotapi.NewKeyboardButton(h.loc.Get(LangKazakh, "kazakh")),
		),
	)
	keyboard.OneTimeKeyboard = true
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

func (h *BotHandler) startRegistration(chatID int64) {
	lang := h.getUserLanguage(chatID)
	h.state[chatID] = "awaiting_name"
	h.temp[chatID] = make(map[string]string)
	h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "registration_required")))
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

func (h *BotHandler) SendNotification(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := h.bot.Send(msg)
	return err
}

// internal/delivery/telegram/bot_handler.go
func (h *BotHandler) sendVideoLink(chatID int64, apptID int) {
	lang := h.getUserLanguage(chatID)
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://telemed-76fw.onrender.com"
	}

	videoURL := fmt.Sprintf("%s/webrtc/room.html?appointment_id=%d&role=patient",
		baseURL, apptID)

	msg := tgbotapi.NewMessage(chatID,
		h.loc.Get(lang, "video_link_message")+"\n"+videoURL)
	h.bot.Send(msg)
}

func (h *BotHandler) SendVideoLink(chatID int64, appointmentID int) error {
	h.sendVideoLink(chatID, appointmentID)
	return nil
}
