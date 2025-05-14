package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
