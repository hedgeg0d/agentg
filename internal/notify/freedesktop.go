package notify

import (
	"strings"
	"sync/atomic"

	"github.com/godbus/dbus/v5"
)

const (
	fdName  = "org.freedesktop.Notifications"
	fdPath  = "/org/freedesktop/Notifications"
	fdIface = "org.freedesktop.Notifications"
)

func exportFreedesktop(conn *dbus.Conn, h Handler) error {
	obj := &freedesktop{handler: h}
	if err := conn.Export(obj, fdPath, fdIface); err != nil {
		return err
	}
	return claim(conn, fdName)
}

type freedesktop struct {
	handler Handler
	counter atomic.Uint32
}

func (f *freedesktop) Notify(appName string, replacesID uint32, appIcon, summary, body string,
	actions []string, hints map[string]dbus.Variant, timeout int32) (uint32, *dbus.Error) {
	title := summary
	if appName != "" {
		title = appName + ": " + summary
	}
	f.handler(strings.TrimSpace(title), body)
	return f.counter.Add(1), nil
}

func (f *freedesktop) CloseNotification(id uint32) *dbus.Error { return nil }

func (f *freedesktop) GetCapabilities() ([]string, *dbus.Error) {
	return []string{"body"}, nil
}

func (f *freedesktop) GetServerInformation() (string, string, string, string, *dbus.Error) {
	return "agentg", "agentg", "1.0", "1.2", nil
}
