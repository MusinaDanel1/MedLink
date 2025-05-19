package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *BotHandler) handleStart(chatID int64) {
	isRegistered := h.patient.Exists(chatID)

	if isRegistered {
		h.sendMainMenu(chatID)
		return
	}

	h.state[chatID] = "awaiting_name"
	h.temp[chatID] = make(map[string]string)
	h.bot.Send(tgbotapi.NewMessage(chatID,
		`Для начала работы необходимо зарегистрироваться.

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
