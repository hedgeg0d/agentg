package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/hedgeg0d/agentg/internal/services"
)

func (b *Bot) showServices(chat int64) {
	names := b.store.Services()
	if len(names) == 0 {
		b.sendInline(chat, "⚙️ *Services*\nNo pinned services yet. Tap Add to track one.", servicesKeyboard(nil))
		return
	}
	var list []services.Status
	for _, n := range names {
		list = append(list, services.Get(n))
	}
	b.sendInline(chat, "⚙️ *Services*", servicesKeyboard(list))
}

func (b *Bot) handleServiceAdd(msg *tgbotapi.Message) {
	chat := msg.Chat.ID
	b.setMode(chat, modeIdle)
	if msg.Text == btnExit {
		b.sendKeyboard(chat, "Cancelled.", mainKeyboard())
		return
	}
	name := normalize(msg.Text)
	st := services.Get(name)
	if !st.Loaded {
		b.send(chat, "⚠️ Service `"+name+"` not found on this system.")
		return
	}
	b.store.AddService(name)
	b.sendInline(chat, renderService(st), serviceKeyboard(name, true))
}

func (b *Bot) serviceCallback(chat int64, cb *tgbotapi.CallbackQuery, rest string) {
	action, name, _ := strings.Cut(rest, ":")
	id := cb.Message.MessageID
	switch action {
	case "list":
		b.editServices(chat, id)
		b.answer(cb.ID, "")
	case "add":
		b.setMode(chat, modeServiceAdd)
		b.sendKeyboard(chat, "Send a service name to track (e.g. `nginx`).", shellKeyboard())
		b.answer(cb.ID, "")
	case "open":
		st := services.Get(name)
		b.editInline(chat, id, renderService(st), serviceKeyboard(name, b.pinned(name)))
		b.answer(cb.ID, "")
	case "start", "stop", "restart":
		b.answer(cb.ID, title(action)+"ing…")
		if err := act(action, name); err != nil {
			b.send(chat, "⚠️ "+err.Error())
		}
		st := services.Get(name)
		b.editInline(chat, id, renderService(st), serviceKeyboard(name, b.pinned(name)))
	case "pin":
		b.store.AddService(name)
		b.editInline(chat, id, renderService(services.Get(name)), serviceKeyboard(name, true))
		b.answer(cb.ID, "Pinned")
	case "unpin":
		b.store.RemoveService(name)
		b.editInline(chat, id, renderService(services.Get(name)), serviceKeyboard(name, false))
		b.answer(cb.ID, "Unpinned")
	}
}

func (b *Bot) editServices(chat int64, id int) {
	names := b.store.Services()
	var list []services.Status
	for _, n := range names {
		list = append(list, services.Get(n))
	}
	text := "⚙️ *Services*"
	if len(list) == 0 {
		text = "⚙️ *Services*\nNo pinned services yet. Tap Add to track one."
	}
	b.editInline(chat, id, text, servicesKeyboard(list))
}

func (b *Bot) pinned(name string) bool {
	for _, n := range b.store.Services() {
		if n == name {
			return true
		}
	}
	return false
}

func act(action, name string) error {
	switch action {
	case "start":
		return services.Start(name)
	case "stop":
		return services.Stop(name)
	case "restart":
		return services.Restart(name)
	}
	return nil
}

func renderService(s services.Status) string {
	return fmt.Sprintf("%s *%s*\nState: `%s` (%s)\nLoaded: `%v`", s.Icon(), s.Name, s.Active, s.Sub, s.Loaded)
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func normalize(name string) string {
	name = strings.TrimSpace(name)
	if !strings.Contains(name, ".") {
		name += ".service"
	}
	return name
}
