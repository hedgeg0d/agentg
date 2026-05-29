package bot

import (
	"os"
	"strings"

	"github.com/hedgeg0d/agentg/internal/monitor"
)

func (b *Bot) sendStatus(chat int64) {
	snap, err := monitor.Sample()
	if err != nil {
		b.send(chat, "⚠️ "+err.Error())
		return
	}
	host, _ := os.Hostname()
	kernel := readTrim("/proc/sys/kernel/osrelease")
	b.send(chat, "🖥 *"+host+"*  `"+kernel+"`\n\n"+snap.Render())
}

func readTrim(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(string(raw))
}
