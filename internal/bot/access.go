package bot

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/hedgeg0d/agentg/internal/auth"
)

func (b *Bot) handleDenied(msg *tgbotapi.Message) {
	if b.auth.BootstrapAdmin(msg.From.ID) {
		b.sendKeyboard(msg.Chat.ID, "🔑 You are the first user — registered as *admin*.", mainKeyboard(true))
		return
	}
	b.send(msg.Chat.ID, "⛔ *Access denied.*\nYour ID: `"+itoa(msg.From.ID)+"`\nUsername: `@"+orDash(msg.From.UserName)+"`\nAsk an admin to grant you access.")
}

func (b *Bot) handlePassword(msg *tgbotapi.Message) {
	chat := msg.Chat.ID
	if b.mode(chat) != modePassword {
		b.setMode(chat, modePassword)
		b.send(chat, "🔒 This bot is password protected.\nSend the access password:")
		return
	}
	exp, ok := b.auth.Authenticate(msg.From.ID, msg.Text)
	if !ok {
		b.send(chat, "❌ Wrong password. Try again:")
		return
	}
	b.setMode(chat, modeIdle)
	note := "Session does not expire."
	if !exp.IsZero() {
		note = "Session valid until " + exp.Format("2006-01-02 15:04") + "."
	}
	b.sendKeyboard(chat, "✅ Access granted. "+note, mainKeyboard(b.auth.IsAdmin(msg.From.ID)))
}

func (b *Bot) usersCommand(msg *tgbotapi.Message) {
	if !b.auth.IsAdmin(msg.From.ID) {
		b.send(msg.Chat.ID, "⛔ Admins only.")
		return
	}
	b.sendInline(msg.Chat.ID, usersText(b.auth.Entries()), usersKeyboard(b.auth.Entries()))
}

func (b *Bot) handleUserAdd(msg *tgbotapi.Message) {
	chat := msg.Chat.ID
	b.setMode(chat, modeIdle)
	if msg.Text == btnExit {
		b.sendKeyboard(chat, "Cancelled.", mainKeyboard(true))
		return
	}
	text := strings.TrimSpace(msg.Text)
	if id, err := strconv.ParseInt(text, 10, 64); err == nil {
		b.auth.AddUser(id)
		b.send(chat, "✅ Added user `"+itoa(id)+"`.")
	} else {
		b.auth.AddUsername(text)
		b.send(chat, "✅ Added user @"+strings.TrimPrefix(text, "@")+".")
	}
	b.sendInline(chat, usersText(b.auth.Entries()), usersKeyboard(b.auth.Entries()))
}

func (b *Bot) userCallback(chat int64, cb *tgbotapi.CallbackQuery, rest string) {
	if !b.auth.IsAdmin(cb.From.ID) {
		b.answer(cb.ID, "Admins only")
		return
	}
	action, arg, _ := strings.Cut(rest, ":")
	switch action {
	case "list":
		b.editInline(chat, cb.Message.MessageID, usersText(b.auth.Entries()), usersKeyboard(b.auth.Entries()))
		b.answer(cb.ID, "")
	case "add":
		b.setMode(chat, modeUserAdd)
		b.sendKeyboard(chat, "Send a Telegram ID or @username to grant access.", shellKeyboard())
		b.answer(cb.ID, "")
	case "rm":
		kind, val, _ := strings.Cut(arg, ":")
		if kind == "id" {
			if id, err := strconv.ParseInt(val, 10, 64); err == nil {
				b.auth.RemoveUser(id)
			}
		} else {
			b.auth.RemoveUsername(val)
		}
		b.editInline(chat, cb.Message.MessageID, usersText(b.auth.Entries()), usersKeyboard(b.auth.Entries()))
		b.answer(cb.ID, "Removed")
	}
}

func usersText(entries []auth.Entry) string {
	if len(entries) == 0 {
		return "👥 *Whitelist*\nNo additional users. Tap Add to grant access."
	}
	return "👥 *Whitelist*\nTap a user to revoke access."
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
