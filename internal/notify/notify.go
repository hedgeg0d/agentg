package notify

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const (
	busName = "com.agentg.Notifier"
	objPath = "/com/agentg/Notifier"
	iface   = "com.agentg.Notifier"
)

type Handler func(title, body string)

type Service struct {
	conn *dbus.Conn
}

func Start(systemBus, replaceNotifySend bool, h Handler) (*Service, error) {
	conn, err := connect(systemBus)
	if err != nil {
		return nil, err
	}
	if err := exportNotifier(conn, h); err != nil {
		conn.Close()
		return nil, err
	}
	if replaceNotifySend {
		if err := exportFreedesktop(conn, h); err != nil {
			log.Printf("notify: not intercepting notify-send: %v", err)
		}
	}
	return &Service{conn: conn}, nil
}

func (s *Service) Close() error {
	return s.conn.Close()
}

func connect(systemBus bool) (*dbus.Conn, error) {
	if systemBus {
		return dbus.ConnectSystemBus()
	}
	return dbus.ConnectSessionBus()
}

func claim(conn *dbus.Conn, name string) error {
	reply, err := conn.RequestName(name, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return fmt.Errorf("dbus: name %s already taken", name)
	}
	return nil
}

func exportNotifier(conn *dbus.Conn, h Handler) error {
	obj := &notifier{handler: h}
	if err := conn.Export(obj, objPath, iface); err != nil {
		return err
	}
	node := &introspect.Node{
		Name: objPath,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			{
				Name: iface,
				Methods: []introspect.Method{{
					Name: "Notify",
					Args: []introspect.Arg{
						{Name: "title", Type: "s", Direction: "in"},
						{Name: "body", Type: "s", Direction: "in"},
					},
				}},
			},
		},
	}
	if err := conn.Export(introspect.NewIntrospectable(node), objPath, "org.freedesktop.DBus.Introspectable"); err != nil {
		return err
	}
	return claim(conn, busName)
}

type notifier struct {
	handler Handler
}

func (n *notifier) Notify(title, body string) *dbus.Error {
	n.handler(title, body)
	return nil
}
