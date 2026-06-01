package bot

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) dispatch(c tgbotapi.Chattable) {
	if _, err := b.api.Send(c); err != nil {
		log.Printf("send: %v", err)
	}
}

func (b *Bot) send(chat int64, text string) {
	msg := tgbotapi.NewMessage(chat, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.dispatch(msg)
}

func (b *Bot) sendPlain(chat int64, text string) {
	b.dispatch(tgbotapi.NewMessage(chat, text))
}

func (b *Bot) sendKeyboard(chat int64, text string, kb tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chat, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = kb
	b.dispatch(msg)
}

func (b *Bot) sendInline(chat int64, text string, kb tgbotapi.InlineKeyboardMarkup) (int, error) {
	msg := tgbotapi.NewMessage(chat, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = kb
	sent, err := b.api.Send(msg)
	if err != nil {
		log.Printf("sendInline: %v", err)
	}
	return sent.MessageID, err
}

func (b *Bot) editInline(chat int64, id int, text string, kb tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageTextAndMarkup(chat, id, text, kb)
	edit.ParseMode = tgbotapi.ModeMarkdown
	b.dispatch(edit)
}

func (b *Bot) answer(cbID, text string) {
	if _, err := b.api.Request(tgbotapi.NewCallback(cbID, text)); err != nil {
		log.Printf("answer: %v", err)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
