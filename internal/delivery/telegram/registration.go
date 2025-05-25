package telegram

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *BotHandler) handleStart(chatID int64) {
	// Always start with language selection
	h.state[chatID] = "choosing_language"
	h.sendLanguageSelection(chatID)
}

func (h *BotHandler) handleUserInput(chatID int64, msg string) {
	lang := h.getUserLanguage(chatID)

	switch h.state[chatID] {
	case "awaiting_name":
		isValid, errorMsg := ValidateFullName(msg, lang)
		if !isValid {
			h.bot.Send(tgbotapi.NewMessage(chatID, errorMsg))
			return
		}

		// Форматируем ФИО
		formattedName := FormatFullName(msg)
		h.temp[chatID]["full_name"] = formattedName
		h.state[chatID] = "awaiting_iin"
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "enter_iin")))

	case "awaiting_iin":
		isValid, errorMsg := ValidateIIN(msg, lang)
		if !isValid {
			h.bot.Send(tgbotapi.NewMessage(chatID, errorMsg))
			return
		}

		fullName := h.temp[chatID]["full_name"]
		iin := strings.TrimSpace(msg)

		err := h.patient.FindOrRegister(chatID, fullName, iin)
		if err != nil {
			h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "registration_error")))
			return
		}

		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "registration_success")))
		delete(h.state, chatID)
		delete(h.temp, chatID)

		h.sendMainMenu(chatID)

	default:
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "invalid_command")))
	}
}
