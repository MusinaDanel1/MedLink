package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *BotHandler) sendMainMenu(chatID int64) {
	lang := h.getUserLanguage(chatID)
	msg := tgbotapi.NewMessage(chatID, h.loc.Get(lang, "main_menu_question"))

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(h.loc.Get(lang, "book_appointment")),
			tgbotapi.NewKeyboardButton(h.loc.Get(lang, "ai_consultation")),
		),
	)
	keyboard.OneTimeKeyboard = false
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}
