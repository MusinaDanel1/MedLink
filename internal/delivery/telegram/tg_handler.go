package telegram

import (
	"telemed/internal/usecase"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot     *tgbotapi.BotAPI
	patient *usecase.PatientService
	state   map[int64]string
	temp    map[int64]map[string]string
}

func NewBotHandler(bot *tgbotapi.BotAPI, patient *usecase.PatientService) *BotHandler {
	return &BotHandler{
		bot:     bot,
		patient: patient,
		state:   make(map[int64]string),
		temp:    make(map[int64]map[string]string),
	}
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
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

func (h *BotHandler) handleStart(chatID int64) {
	isRegistered := h.patient.Exists(chatID)

	if isRegistered {
		h.sendMainMenu(chatID)
		return
	}

	h.state[chatID] = "awaiting_name"
	h.temp[chatID] = make(map[string]string)
	h.bot.Send(tgbotapi.NewMessage(chatID,
		`Добро пожаловать в Телемед-бот! 🤖

Этот бот поможет вам:
• 📅 записаться на прием к врачу
• 💬 получить консультацию у ИИ

Для начала работы необходимо зарегистрироваться.

Пожалуйста, введите ваше ФИО:`))
}

func (h *BotHandler) handleUserInput(chatID int64, msg string) {
	switch h.state[chatID] {
	case "awaiting_name":
		h.temp[chatID]["full_name"] = msg
		h.state[chatID] = "awaiting_iin"
		h.bot.Send(tgbotapi.NewMessage(chatID, "Введите ваш ИИН:"))

	case "awaiting_iin":
		fullName := h.temp[chatID]["full_name"]
		iin := msg

		err := h.patient.FindOrRegister(chatID, fullName, iin)
		if err != nil {
			h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка регистрации. Попробуйте позже."))
			return
		}

		h.bot.Send(tgbotapi.NewMessage(chatID, "Вы успешно зарегистрированы!"))
		delete(h.state, chatID)
		delete(h.temp, chatID)

		h.sendMainMenu(chatID)

	default:
		h.bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, выберите действие из меню или введите /start для начала."))
	}
}

func (h *BotHandler) sendMainMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Что вы хотите сделать?")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Записаться к врачу", "book_appointment"),
			tgbotapi.NewInlineKeyboardButtonData("💬 Консультация с ИИ", "ai_consultation"),
		),
	)
	h.bot.Send(msg)
}
