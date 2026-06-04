package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/hedgeg0d/agentg/internal/power"
)

func (b *Bot) powerCommand(msg *tgbotapi.Message) {
	if !b.auth.IsAdmin(msg.From.ID) {
		b.send(msg.Chat.ID, "⛔ Admins only.")
		return
	}
	b.sendInline(msg.Chat.ID, "🔌 *Power*\nChoose an action.", powerMenuKeyboard())
}

func (b *Bot) powerCallback(chat int64, cb *tgbotapi.CallbackQuery, rest string) {
	if !b.auth.IsAdmin(cb.From.ID) {
		b.answer(cb.ID, "Admins only")
		return
	}
	id := cb.Message.MessageID
	switch rest {
	case "menu":
		b.editInline(chat, id, "🔌 *Power*\nChoose an action.", powerMenuKeyboard())
		b.answer(cb.ID, "")
	case "ask:reboot":
		b.editInline(chat, id, "⚠️ *Reboot this machine?*", powerConfirmKeyboard("reboot", "Reboot"))
		b.answer(cb.ID, "")
	case "ask:poweroff":
		b.editInline(chat, id, "⚠️ *Power off this machine?*", powerConfirmKeyboard("poweroff", "Power off"))
		b.answer(cb.ID, "")
	case "do:reboot":
		b.runPower(chat, id, cb, "♻️ Rebooting now…", power.Reboot)
	case "do:poweroff":
		b.runPower(chat, id, cb, "🔌 Powering off now…", power.Poweroff)
	}
}

func (b *Bot) runPower(chat int64, msgID int, cb *tgbotapi.CallbackQuery, notice string, action func() error) {
	b.answer(cb.ID, "Executing")
	b.editInline(chat, msgID, notice, emptyInline())
	if err := action(); err != nil {
		b.send(chat, "⚠️ "+err.Error())
	}
}
