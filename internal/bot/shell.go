package bot

import (
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/hedgeg0d/agentg/internal/shell"
)

const (
	tgLimit      = 3500
	editInterval = 2 * time.Second
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
		b.sendKeyboard(chat, "Shell closed.", mainKeyboard(b.auth.IsAdmin(msg.From.ID)))
		return
	}
	ch, err := b.shell.Run(chat, msg.Text)
	if err != nil {
		b.shell.Reset(chat)
		b.send(chat, "⚠️ "+err.Error())
		return
	}
	b.streamOutput(chat, ch, b.cfg.Shell.Streaming())
}

func (b *Bot) streamOutput(chat int64, ch <-chan shell.Update, live bool) {
	var msgID int
	var shown string
	var lastEdit time.Time

	for u := range ch {
		if u.Done {
			b.finishOutput(chat, msgID, u, live)
			return
		}
		if !live {
			continue
		}
		if msgID == 0 {
			msgID, _ = b.sendMarkdownV2(chat, renderRun("⏳ running", u.Output))
			shown, lastEdit = u.Output, time.Now()
			continue
		}
		if u.Output != shown && time.Since(lastEdit) >= editInterval {
			b.editMarkdownV2(chat, msgID, renderRun("⏳ running", u.Output))
			shown, lastEdit = u.Output, time.Now()
		}
	}
}

func (b *Bot) finishOutput(chat int64, msgID int, u shell.Update, live bool) {
	status := "✅ done"
	if u.Err != nil {
		status = "⚠️ " + u.Err.Error()
	}
	if live && msgID != 0 {
		b.editMarkdownV2(chat, msgID, renderRun(status, u.Output))
		if len(u.Output) > tgLimit {
			b.sendCode(chat, u.Output)
		}
		return
	}
	text := u.Output
	if u.Err != nil {
		text = strings.TrimRight(text+"\n["+u.Err.Error()+"]", "\n")
	}
	b.sendCode(chat, text)
}

func renderRun(status, text string) string {
	return codeBlock(status + "\n\n" + tailLimit(text))
}

func tailLimit(s string) string {
	if len(s) <= tgLimit {
		return s
	}
	return "…" + s[len(s)-tgLimit:]
}

func codeBlock(text string) string {
	if text == "" {
		text = "(no output)"
	}
	esc := strings.NewReplacer("\\", "\\\\", "`", "\\`").Replace(text)
	return "```\n" + esc + "\n```"
}

func (b *Bot) sendCode(chat int64, text string) {
	if text == "" {
		text = "(no output)"
	}
	for len(text) > 0 {
		chunk := text
		if len(chunk) > tgLimit {
			chunk = chunk[:tgLimit]
		}
		text = text[len(chunk):]
		b.sendMarkdownV2(chat, codeBlock(chunk))
	}
}
