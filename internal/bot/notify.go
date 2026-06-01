package bot

import "strings"

func (b *Bot) Notify(title, body string) {
	text := "🔔 " + strings.TrimSpace(title)
	if body = strings.TrimSpace(body); body != "" {
		text += "\n" + body
	}
	for _, id := range b.auth.Recipients() {
		b.sendPlain(id, text)
	}
}
