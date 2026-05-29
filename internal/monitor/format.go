package monitor

import (
	"fmt"
	"strings"
	"time"
)

func (s Snapshot) Render() string {
	var b strings.Builder
	b.WriteString("📊 *System Monitor*\n\n")
	fmt.Fprintf(&b, "CPU   %s `%5.1f%%`\n", bar(s.CPU), s.CPU)
	fmt.Fprintf(&b, "RAM   %s `%s`\n", bar(pct(s.MemUsed, s.MemTotal)), size(s.MemUsed)+"/"+size(s.MemTotal))
	if s.SwapTotal > 0 {
		fmt.Fprintf(&b, "Swap  %s `%s`\n", bar(pct(s.SwapUsed, s.SwapTotal)), size(s.SwapUsed)+"/"+size(s.SwapTotal))
	}
	fmt.Fprintf(&b, "Disk  %s `%s`\n", bar(pct(s.DiskUsed, s.DiskTotal)), size(s.DiskUsed)+"/"+size(s.DiskTotal))
	fmt.Fprintf(&b, "\nLoad `%s`  Up `%s`\n", s.Load, uptime(s.Uptime))
	fmt.Fprintf(&b, "_updated %s_", time.Now().Format("15:04:05"))
	return b.String()
}

func pct(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}

func bar(p float64) string {
	const n = 10
	filled := int(p/100*n + 0.5)
	if filled > n {
		filled = n
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", n-filled)
}

func size(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func uptime(d time.Duration) string {
	d = d.Round(time.Minute)
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, h, m)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}
