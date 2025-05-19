package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *BotHandler) sendMainMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Что вы хотите сделать?")

	// Create keyboard with two buttons
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 Записаться к врачу"),
			tgbotapi.NewKeyboardButton("💬 Консультация с ИИ"),
		),
	)
	keyboard.OneTimeKeyboard = false // Keep keyboard visible
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}
