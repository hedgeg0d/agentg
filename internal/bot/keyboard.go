package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/hedgeg0d/agentg/internal/services"
)

const (
	btnShell    = "💻 Shell"
	btnMonitor  = "📊 Monitor"
	btnServices = "⚙️ Services"
	btnStatus   = "ℹ️ Status"
	btnExit     = "⬅️ Back"
)

func mainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnShell),
			tgbotapi.NewKeyboardButton(btnMonitor),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnServices),
			tgbotapi.NewKeyboardButton(btnStatus),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func shellKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnExit),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func monitorKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "mon:refresh"),
			tgbotapi.NewInlineKeyboardButtonData("⏹ Stop", "mon:stop"),
		),
	)
}

func servicesKeyboard(list []services.Status) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, s := range list {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(s.Icon()+" "+s.Name, "svc:open:"+s.Name),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ Add", "svc:add"),
		tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "svc:list"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func serviceKeyboard(name string, pinned bool) tgbotapi.InlineKeyboardMarkup {
	pin := tgbotapi.NewInlineKeyboardButtonData("📌 Pin", "svc:pin:"+name)
	if pinned {
		pin = tgbotapi.NewInlineKeyboardButtonData("📍 Unpin", "svc:unpin:"+name)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶️ Start", "svc:start:"+name),
			tgbotapi.NewInlineKeyboardButtonData("⏹ Stop", "svc:stop:"+name),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Restart", "svc:restart:"+name),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔃 Refresh", "svc:open:"+name),
			pin,
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", "svc:list"),
		),
	)
}
