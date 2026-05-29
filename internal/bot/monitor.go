package bot

import (
	"time"

	"github.com/hedgeg0d/agentg/internal/monitor"
)

func (b *Bot) startMonitor(chat int64) {
	b.stopMonitor(chat)
	stop := make(chan struct{})
	b.mu.Lock()
	b.monitors[chat] = stop
	b.mu.Unlock()

	snap, err := monitor.Sample()
	if err != nil {
		b.send(chat, "⚠️ "+err.Error())
		return
	}
	id, err := b.sendInline(chat, snap.Render(), monitorKeyboard())
	if err != nil {
		return
	}
	go b.monitorLoop(chat, id, stop)
}

func (b *Bot) monitorLoop(chat int64, msgID int, stop chan struct{}) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			snap, err := monitor.Sample()
			if err != nil {
				continue
			}
			b.editInline(chat, msgID, snap.Render(), monitorKeyboard())
		}
	}
}

func (b *Bot) stopMonitor(chat int64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if stop, ok := b.monitors[chat]; ok {
		close(stop)
		delete(b.monitors, chat)
		return true
	}
	return false
}
