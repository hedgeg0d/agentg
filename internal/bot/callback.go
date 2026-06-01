package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/hedgeg0d/agentg/internal/monitor"
)

func emptyInline() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}}
}

func (b *Bot) routeCallback(cb *tgbotapi.CallbackQuery) {
	chat := cb.Message.Chat.ID
	domain, rest, _ := strings.Cut(cb.Data, ":")
	switch domain {
	case "mon":
		b.monitorCallback(chat, cb, rest)
	case "svc":
		b.serviceCallback(chat, cb, rest)
	case "usr":
		b.userCallback(chat, cb, rest)
	default:
		b.answer(cb.ID, "")
	}
}

func (b *Bot) monitorCallback(chat int64, cb *tgbotapi.CallbackQuery, action string) {
	switch action {
	case "stop":
		b.stopMonitor(chat)
		b.editInline(chat, cb.Message.MessageID, cb.Message.Text+"\n\n⏹ _stopped_", emptyInline())
		b.answer(cb.ID, "Stopped")
	case "refresh":
		if snap, err := monitor.Sample(); err == nil {
			b.editInline(chat, cb.Message.MessageID, snap.Render(), monitorKeyboard())
		}
		b.answer(cb.ID, "")
	}
}
