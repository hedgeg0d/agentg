package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) enterShell(chat int64) {
	b.setMode(chat, modeShell)
	b.sendKeyboard(chat, "💻 *Shell mode*\nSession persists cwd and env. Send commands; tap Back to exit.", shellKeyboard())
}

func (b *Bot) handleShellInput(msg *tgbotapi.Message) {
	chat := msg.Chat.ID
	if msg.Text == btnExit {
		b.setMode(chat, modeIdle)
		b.shell.Reset(chat)
		b.sendKeyboard(chat, "Shell closed.", mainKeyboard())
		return
	}
	out, err := b.shell.Run(chat, msg.Text)
	if err != nil {
		b.shell.Reset(chat)
		b.send(chat, "⚠️ "+err.Error()+"\n_session restarted_")
		return
	}
	b.sendCode(chat, out)
}

func (b *Bot) sendCode(chat int64, text string) {
	if text == "" {
		text = "(no output)"
	}
	const limit = 3900
	for len(text) > 0 {
		chunk := text
		if len(chunk) > limit {
			chunk = chunk[:limit]
		}
		text = text[len(chunk):]
		chunk = strings.NewReplacer("\\", "\\\\", "`", "\\`").Replace(chunk)
		msg := tgbotapi.NewMessage(chat, "```\n"+chunk+"\n```")
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		b.api.Send(msg)
	}
}
