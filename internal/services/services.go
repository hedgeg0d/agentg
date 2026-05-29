package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Status struct {
	Name   string
	Active string
	Sub    string
	Loaded bool
}

func (s Status) Icon() string {
	switch s.Active {
	case "active":
		return "🟢"
	case "failed":
		return "🔴"
	case "inactive":
		return "⚪"
	default:
		return "🟡"
	}
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "systemctl", args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func Get(name string) Status {
	s := Status{Name: name}
	out, _ := run("show", name, "--property=ActiveState,SubState,LoadState")
	for _, line := range strings.Split(out, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch k {
		case "ActiveState":
			s.Active = v
		case "SubState":
			s.Sub = v
		case "LoadState":
			s.Loaded = v == "loaded"
		}
	}
	return s
}

func Start(name string) error   { return action("start", name) }
func Stop(name string) error    { return action("stop", name) }
func Restart(name string) error { return action("restart", name) }

func action(verb, name string) error {
	out, err := run(verb, name)
	if err != nil {
		if out == "" {
			return err
		}
		return fmt.Errorf("%s", out)
	}
	return nil
}
