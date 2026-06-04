package bot

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) favCommand(msg *tgbotapi.Message) {
	chat := msg.Chat.ID
	if args := strings.TrimSpace(msg.CommandArguments()); args != "" {
		if b.store.AddCommand(args) {
			b.send(chat, "⭐ Added to favorites: `"+args+"`")
		} else {
			b.send(chat, "Already in favorites.")
		}
		return
	}
	b.showFavorites(chat)
}

func (b *Bot) showFavorites(chat int64) {
	cmds := b.store.Commands()
	if len(cmds) == 0 {
		b.send(chat, "No favorite commands yet. Add one with `/fav <command>`.")
		return
	}
	b.sendInline(chat, favText(), favManageKeyboard(cmds))
}

func (b *Bot) commandCallback(chat int64, cb *tgbotapi.CallbackQuery, rest string) {
	action, arg, _ := strings.Cut(rest, ":")
	switch action {
	case "run":
		i, err := strconv.Atoi(arg)
		if err != nil {
			b.answer(cb.ID, "")
			return
		}
		cmd, ok := b.store.CommandAt(i)
		if !ok {
			b.answer(cb.ID, "Gone")
			return
		}
		b.answer(cb.ID, "▶️ "+cmdLabel(cmd))
		b.runShellCommand(chat, cmd)
	case "rm":
		if i, err := strconv.Atoi(arg); err == nil {
			b.store.RemoveCommandAt(i)
		}
		cmds := b.store.Commands()
		if len(cmds) == 0 {
			b.editInline(chat, cb.Message.MessageID, "No favorite commands.", emptyInline())
		} else {
			b.editInline(chat, cb.Message.MessageID, favText(), favManageKeyboard(cmds))
		}
		b.answer(cb.ID, "Removed")
	}
}

func favText() string {
	return "⭐ *Favorite commands*\nTap a command to remove it."
}
