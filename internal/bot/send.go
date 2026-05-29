package bot

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) send(chat int64, text string) {
	msg := tgbotapi.NewMessage(chat, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(msg)
}

func (b *Bot) sendKeyboard(chat int64, text string, kb tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chat, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = kb
	b.api.Send(msg)
}

func (b *Bot) sendInline(chat int64, text string, kb tgbotapi.InlineKeyboardMarkup) (int, error) {
	msg := tgbotapi.NewMessage(chat, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = kb
	sent, err := b.api.Send(msg)
	return sent.MessageID, err
}

func (b *Bot) editInline(chat int64, id int, text string, kb tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageTextAndMarkup(chat, id, text, kb)
	edit.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(edit)
}

func (b *Bot) answer(cbID, text string) {
	b.api.Request(tgbotapi.NewCallback(cbID, text))
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
